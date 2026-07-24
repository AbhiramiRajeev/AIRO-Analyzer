# AIRO-Analyzer

###  Analyzer Event Processing Flow

1️ **Event arrives from Kafka**  
2️ **Add to Redis sorted set**  
3️ **Trim old timestamps**  
4️ **Run sliding window check**  
5️ **Run suspicious IP check**  
6️ **Run impossible travel check**  
7️ **Run device fingerprint check** *(optional)*  
8️ **Combine results** → If **any check is flagged**, **create incident**  
9️ **Insert incident into Postgres** + **publish to `incident_events` Kafka topic**


