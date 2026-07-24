# AIRO-Analyzer

Real-time authentication event analysis engine. Consumes login events from Kafka, runs detection algorithms against each event, and persists incidents to PostgreSQL when threats are identified.

## Architecture

```text
                              +-------------------------+
                              |     Event Producers     |
                              |    (Login Service)      |
                              +------------+------------+
                                           |
                                           | Publish Events
                                           ▼
                              +-------------------------+
                              |         Kafka           |
                              |       auth_events       |
                              |         :9092           |
                              +------------+------------+
                                           |
                                           | Consume Events
                                           ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                            AIRO Analyzer Service                             │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                        Kafka Consumer Group                            │  │
│  └───────────────────────────────┬────────────────────────────────────────┘  │
│                                  │                                           │
│                                  ▼                                           │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                       Analyzer Engine                                  │  │
│  │                                                                        │  │
│  │   • Sliding Window Detection                                           │  │
│  │   • Suspicious IP Detection                                            │  │
│  │   • (Future) Impossible Travel Detection                               │  │
│  │   • (Future) Device Fingerprint Analysis                               │  │
│  └───────────────┬───────────────────────┬───────────────────────┬────────┘  │
│                  │                       │                       │           │
│                  ▼                       ▼                       ▼           │
│        +----------------+      +----------------+      +----------------+    │
│        |     Redis      |      |   PostgreSQL   |      | Kafka Producer |    │
│        |     :6379      |      |     :5432      |      | incident_events|    │
│        +----------------+      +----------------+      +--------+-------+    │
└──────────────────────────────────────────────────────────────────────────────┘
                                   │                              │
                                   │                              │
                                   ▼                              ▼
                      +---------------------------+     +----------------------+
                      |   incidents table         |     |  incident_events     |
                      |      PostgreSQL           |     |      Kafka Topic     |
                      +---------------------------+     +----------------------+
```

### Component Overview

| Component | Role |
|-----------|------|
| **Kafka Consumer** | Reads events from `auth_events` topic, deserializes JSON, dispatches to analyzer |
| **Analyzer Service** | Runs detection algorithms, decides whether to create an incident |
| **Redis** | Stores sliding window of failed login timestamps and suspicious IP set |
| **PostgreSQL** | Persists confirmed incidents to the `incidents` table |
| **Kafka Producer** | Publishes detected incidents to the `incident_events` topic for downstream consumers |

## Event Schema

Events are consumed from Kafka as JSON with this structure:

```json
{
  "event_type": "login",
  "user_id": "alice",
  "ip_address": "192.168.1.100",
  "status": "failure",
  "timestamp": "2026-01-15T10:30:00Z"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `event_type` | string | yes | Type of authentication event (e.g. `login`, `mfa_verify`) |
| `user_id` | string | yes | Unique identifier for the user |
| `ip_address` | string | yes | Source IP address of the request |
| `status` | string | yes | Outcome of the event: `success` or `failure` |
| `timestamp` | string | yes | ISO 8601 timestamp of when the event occurred |

### Incident Schema

When a threat is detected, an incident is created:

```json
{
  "user_id": "alice",
  "ip_address": "192.168.1.100",
  "event_type": "login",
  "timestamp": "2026-01-15T10:30:05Z",
  "details": "Multiple failed login attempts detected"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | User associated with the incident |
| `ip_address` | string | Source IP address |
| `event_type` | string | Event type that triggered the incident |
| `timestamp` | string | Time the incident was created |
| `details` | string | Human-readable description of the detected threat |

## Detection Algorithms

### 1. Sliding Window (Brute-Force Detection)

**Purpose:** Detects repeated failed authentication attempts within a configurable time window.

**How it works:**

1. When an event has `status: "failure"`, the current timestamp is added to a Redis sorted set keyed by `failure-{user_id}`.
2. Old entries outside the sliding window (`WINDOW_SIZE` seconds) are pruned using `ZREMRANGEBYSCORE`.
3. The remaining count is retrieved via `ZCOUNT`.
4. If the count >= `FAILURE_THRESHOLD`, an incident is created.

**Redis data structure:** Sorted set where members are timestamps and scores are timestamps.

**Configuration:**
- `WINDOW_SIZE` — width of the sliding window in seconds (default: `300`)
- `FAILURE_THRESHOLD` — number of failures before triggering an incident (default: `5`)

**Example:** With `WINDOW_SIZE=300` and `FAILURE_THRESHOLD=3`, if user `alice` fails to log in 3 times within 5 minutes, an incident is created.

### 2. Suspicious IP Detection

**Purpose:** Flags events originating from known malicious IP addresses.

**How it works:**

1. A set of known-bad IPs is maintained in Redis under the key `suspicious_ips`.
2. For every incoming event (regardless of status), the source IP is checked against this set using `SISMEMBER`.
3. If the IP is in the set, an incident is created.

**Redis data structure:** Set (`suspicious_ips`) containing known malicious IP addresses.

**Populating the blocklist:** Use the `AddSuspiciousIp` method or directly add IPs to Redis:
```
SADD suspicious_ips "1.2.3.4" "5.6.7.8"
```

### 3. Impossible Travel Detection (Planned)

**Purpose:** Would detect logins from geographically distant locations in an impossibly short time.

**Status:** Not yet implemented. The architecture supports adding this as a third check in the analyzer pipeline.

## Project Structure

```
AIRO-Analyzer/
├── cmd/
│   └── main.go              # Application entrypoint
├── config/
│   ├── config.go            # Configuration loading (env vars / YAML)
│   └── config_test.go
├── internal/
│   ├── analyzer/
│   │   ├── analyzer.go      # Core detection logic
│   │   └── analyzer_test.go
│   ├── db/
│   │   └── db.go            # PostgreSQL repository
│   ├── kafka/
│   │   ├── consumer.go      # Kafka consumer group
│   │   ├── producer.go      # Kafka producer for incidents
│   │   └── kafka_test.go
│   ├── models/
│   │   ├── models.go        # Event and Incident structs
│   │   └── models_test.go
│   └── redis/
│       └── redis.go         # Redis client (sorted sets, sets)
├── .github/
│   └── workflows/
│       └── ci.yml           # GitHub Actions CI pipeline
├── docker-compose.yml       # Full local stack
├── Dockerfile               # Multi-stage Go build
├── go.mod
└── go.sum
```

## Configuration

Configuration is loaded via environment variables or a YAML file (`config.yaml` in the working directory or `.config/`).

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `KAFKA_BROKERS` | list | — | Comma-separated list of Kafka broker addresses |
| `KAFKA_TOPIC` | string | — | Kafka topic to consume events from |
| `INCIDENT_TOPIC` | string | — | Kafka topic to publish incidents to |
| `KAFKA_CONSUMER_GROUP` | string | — | Consumer group ID |
| `WINDOW_SIZE` | int | `300` | Sliding window size in seconds |
| `FAILURE_THRESHOLD` | int | `5` | Failed attempts before triggering an incident |
| `POSTGRES_HOST` | string | — | PostgreSQL host |
| `POSTGRES_PORT` | int | `5432` | PostgreSQL port |
| `POSTGRES_USERNAME` | string | — | PostgreSQL username |
| `POSTGRES_PASSWORD` | string | — | PostgreSQL password |
| `POSTGRES_DATABASE` | string | — | PostgreSQL database name |
| `POSTGRES_SSL_MODE` | string | `disable` | PostgreSQL SSL mode |
| `REDIS_ADDRESS` | string | — | Redis address (host:port) |
| `REDIS_PASSWORD` | string | `""` | Redis password (empty if none) |

### Example YAML config

```yaml
KAFKA_BROKERS:
  - localhost:9092
KAFKA_TOPIC: auth_events
INCIDENT_TOPIC: incident_events
KAFKA_CONSUMER_GROUP: airo-analyzer-group
WINDOW_SIZE: 300
FAILURE_THRESHOLD: 5
POSTGRES_HOST: localhost
POSTGRES_PORT: 5432
POSTGRES_USERNAME: airo_user
POSTGRES_PASSWORD: airo_password
POSTGRES_DATABASE: airo_db
POSTGRES_SSL_MODE: disable
REDIS_ADDRESS: localhost:6379
REDIS_PASSWORD: ""
```

## Setup

### Prerequisites

- Go 1.24+
- Docker and Docker Compose (for full stack)

### Quick Start (Docker Compose)

```bash
docker compose up --build
```

This starts Kafka, Zookeeper, PostgreSQL, Redis, and the analyzer service. All services are health-checked before the app starts.

### Local Development

1. Start infrastructure:

```bash
docker compose up -d zookeeper kafka postgres redis
```

2. Set environment variables (or create `config.yaml`):

```bash
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=auth_events
export INCIDENT_TOPIC=incident_events
export KAFKA_CONSUMER_GROUP=airo-analyzer-group
export WINDOW_SIZE=300
export FAILURE_THRESHOLD=5
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USERNAME=airo_user
export POSTGRES_PASSWORD=airo_password
export POSTGRES_DATABASE=airo_db
export POSTGRES_SSL_MODE=disable
export REDIS_ADDRESS=localhost:6379
export REDIS_PASSWORD=""
```

3. Run:

```bash
go run ./cmd/main.go
```

### Creating the Kafka topic

```bash
docker compose exec kafka kafka-topics \
  --create --topic auth_events \
  --bootstrap-server localhost:9092 \
  --partitions 3 --replication-factor 1
```

## Testing

```bash
go test ./... -v
```

Tests use mock implementations of Redis and PostgreSQL interfaces — no external services required.

## Database Schema

The `incidents` table is created automatically on startup:

```sql
CREATE TABLE IF NOT EXISTS incidents (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    ip VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    incident_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## License

MIT
