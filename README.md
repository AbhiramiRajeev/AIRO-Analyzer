# AIRO-Analyzer

### 🔁 Analyzer Event Processing Flow

1️⃣ **Event arrives from Kafka**  
2️⃣ **Add to Redis sorted set**  
3️⃣ **Trim old timestamps**  
4️⃣ **Run sliding window check**  
5️⃣ **Run suspicious IP check**  
6️⃣ **Run impossible travel check**  
7️⃣ **Run device fingerprint check** *(optional)*  
8️⃣ **Combine results** → If **any check is flagged**, **create incident**  
9️⃣ **Insert incident into Postgres** + **publish to `incident_events` Kafka topic**

Producer = Just send a message to a topic.
Consumer = Join a group, subscribe to topic(s), handle rebalancing, handle partitions, etc.

Kafka consumers in Go (with Sarama) need to implement an interface (ConsumerGroupHandler) with multiple lifecycle methods.

[ Kafka ConsumerHandler ]
         |
         V
[ analyzer.Service.AnalyzeEvent() ]
         |
     +---+----+-------------------------+
     |        |                         |
 Redis   Postgres               Kafka Producer (for publishing)


### Kafka ConsumerGroupHandler Methods

| Method           | Purpose                                                                 | Required?                     |
|----------------  |-------------------------------------------------------------------------|-------------------------------|
| `Setup()`        | Called once per partition before consuming starts                       | Optional — can return `nil`   |
| `Cleanup()`      | Called after consuming stops (e.g., rebalance, shutdown)                | Optional — can return `nil`   |
| `ConsumeClaim()` | Where you actually consume messages from the Kafka topic                | ✅ Must be implemented        |
