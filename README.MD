# Voucher Manager

# Architecture
Service will listen through HTTP API for voucher redemption requests. Redeeming is done atomically to avoid inconsistency. 


```text
                                                                                           ┌───SQL-Database──┐
                                                                                           │ Highly Available│
External Network                   Internal Network                                        │ Replicated      │
                │                                                                          │                 │
                │                                                                          │      ┌────────┐ │
     ┌───────┐  │      ┌───Wallet──────────────────────────────┐                           │      │        │ │
     │Client │  │      │                                       │                           │    ┌─┼──────┐ │ │
     │       │  │      │                                       │                           │    │ │      │ │ │
     │       │  │      │   ┌────────────┬───Application────┐   │                           │  ┌─┼─┼────┐ │ │ │
     │       ◄──┼─REST─┼───►   REST API │                  │   │              ┌────────────►  │ │ │    │ │ │ │
     │       │  │      │   │            │                  │   │              │            │  │ │ └────┼─┼─┘ │
     └───────┘  │      │   │            │                  │   │              │            │  │ │      │ │   │
                │      │   │            │    Monotiring    ◄───┼──────TCP─────┘            │  │ └──────┼─┘   │
                │      │   │            │                  │   │                           │  │        │     │
                │      │   │            │    Repository    │   │                           │  └────────┘     │
                │      │   │            │                  │   │                 ┌─────────►                 │
                │      │   └─────▲──────┴──────────────────┘   │                 │         │                 │
                │      │         │                             │                 │         │                 │
                │      └─────────┼─────────────────────────────┘       ┌─TCP─────┘         └─────────────────┘
                │                │                                     │
                │                │                                     │         ┌─REDIS-Sentinel─┐
                │                └─────REST─────────┐                  │         │                │
                │                                   │                  │         │       ┌──────┐ │
                │      ┌──Voucher───────────────────┼──────────┐       │         │       │      │ │
    ┌────────┐  │      │                            │          │       │         │    ┌──┼───┐  │ │
    │ Client │  │      │                            │          │       │         │    │  │   │  │ │
    │        │  │      │    ┌─────────┬─────────────▼─────┐    │       │         │  ┌─┼──┼─┐ │  │ │
    │        │  │      │    │ REST API│     Application   ◄────┼───────┘         │  │ │  │ │ │  │ │
    │        ◄──┼──────┼────►         │                   │    │                 │  │ │  └─┼─┼──┘ │
    │        │  │      │    │         │                   ◄────┼────TCP──────────►  │ │    │ │    │
    │        │  │      │    │         │                   │    │                 │  │ └────┼─┘    │
    └────────┘  │      │    │         │                   │    │                 │  │      │      │
                │      │    └─────────┴───────────────────┘    │                 │  └──────┘      │
                │      │                                       │                 │                │
                │      └───────────────────────────────────────┘                 └────────────────┘
                │
                │
                │
                │
```

### Redis
Redis is used as an in memory lock mechanism to handle high load key contentions. Redis setup is considered to have :
- High availability guarantees via Redis-Sentinel
  - Using sentinel we can't have load sharding and data partitioning schemes hence, **we can't scale redis cluster out easily**. One possible solution is to use Redis cluster but for sake of simplicity we consider Sentinel.
- Resident in private network with reserved bandwidth
- Periodic disk persistence

Redis is a key part of the system and any failure in connection or cluster (network failure) will bring the whole system down.

### SQL Database
[CockraochDB](https://www.cockroachlabs.com/) is used as SQL database setup. CockroachDB is a **cloud native**, **distributed** DBMS with postgres-like schematics.
Cockroach provides High-Availability and tolerance to multi-node failures.

Setting up an Highly-Available cluster of Postgresql is tricky, and it's maintenance would require extra effort. But any HA setup of postgresql is fully supported in application.

### Wallet
Voucher <-> Wallet communication is done via HTTP REST APIs. any failure in the connection will cause the whole system to go down.

### Application
Following technologies are used in application:
- HTTP Server: **gofiber**
- Database ORM: **GORM**
- Metrics: **Prometheus**

Both Voucher and Wallet applications are designed and implemented to scale per load. But Redis/SQL connection limits must be considered and tweaked to support autoscaling completely.

#### API
A Postman collection of all possible API calls are available in [challenge.json](challange.json).

#### Voucher redemption flow
1. In Memory Cache Check
   1. Cache is updated periodically. Any Update in vouchers will need `CacheDuration` time to take effect across all instances.
2. Redis Check (**Atomic**)
   1. Each voucher is represented with a key in redis with `value=voucher_limit`
   2. Each request is followed by a `Decr` redis command to check voucher availability.
   3. Any failure in Redemption will trigger a **revert `Incr` command** to the corresponding key.
3. Apply voucher transaction to wallet service (**failure from this stage forward will cause Revert**)
4. Add voucher redemption to Database.

## Alternative Considerations
### No Redis
we could skip using Redis as a temporary storage altogether. But we need another distributed lock mechanism to keep atomicity of record updates. SQL Database Transactions could be possible but this would add to both system complexity and latency.

**If we can tolerate approximation in voucher usage times**. we can use load balancing with replica syncing in any distributed key-value store.

### Voucher <-> Wallet communication
Intra-service communication could be done using CMQ services (i.e., NATS, Kafka). This approach can help total system availability but can break data consistency in some scenarios.

