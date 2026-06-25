Review our current monolith and propose a microservices decomposition strategy.

Current state:
- Rails monolith, 180k LOC, PostgreSQL, Sidekiq for background jobs
- 3 bounded contexts clearly visible: Orders, Inventory, Notifications
- 8 engineers across 3 teams (each owns one context)
- Deployments are becoming painful — 40min CI, teams blocking each other
- DB has 280 tables, heavy cross-context joins especially Orders ↔ Inventory

Goals:
- Independent deploy per team by Q4 2025
- Keep PostgreSQL (no NoSQL)
- Zero downtime migration — can't big-bang rewrite
- Current traffic: 12k req/min peak

Questions to answer:
1. Which service to extract first (lowest risk, highest independence)?
2. How to handle the Orders ↔ Inventory joins without distributed joins?
3. Strangler fig vs parallel run approach?
4. Data ownership boundaries — shared tables?
