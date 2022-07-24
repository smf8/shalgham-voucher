# Wallet Manager

## Architecture
Service will listen to confirmed transactions through CMQ (It's better to be a persisted solution to provider high availability).


```text
┌─────────┐       ┌──Wallet──────────────────────────────────────────────────────────────────┐
│Message  │       │                                                                          │
│Queue    │       │                                                                          │
│         │       │             ┌───────────┬───Application────┐      ┌───SQL-Database──┐    │
│         │       │             │  Consumer │                  │      │                 │    │
│         │       │             │           │                  │      │                 │    │
│         ├───────┼────Consume──►           │                  │      │                 │    │
│  NATS   │       │             │           │    Monotiring    │      │                 │    │
│         │       │             │           │                  │      │                 │    │
│  KAFKA  │       │             │           │    Repository    │      │                 │    │
│         │       │             │           │                  │      │                 │    │
└─────────┘       │             ├───────────┤                  │      │                 │    │
                  │             │  REST API │                  ◄──────►                 │    │
                  │             │           │                  │      │                 │    │
┌─────────┐       │             │           │                  │      │                 │    │
│ Client  │       │             │           │                  │      │                 │    │
│         │       │             │           │                  │      │                 │    │
│         ◄───────┼─────REST────►           │                  │      │                 │    │
│         │       │             │           │                  │      │                 │    │
│         │       │             │           │                  │      │                 │    │
│         │       │             └───────────┴──────────────────┘      └─────────────────┘    │
│         │       │                                                                          │
│         │       └──────────────────────────────────────────────────────────────────────────┘
│         │
└─────────┘
```