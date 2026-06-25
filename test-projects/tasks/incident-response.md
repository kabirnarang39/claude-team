P0 INCIDENT — Production payments down.

Timeline:
- 14:23 UTC: Alerts fire — payment success rate drops from 99.1% to 0%
- 14:25 UTC: DB CPU at 100%, connection pool exhausted (max 200)
- 14:26 UTC: Slow query log shows: SELECT * FROM transactions WHERE user_id=? takes 45s
- 14:28 UTC: transactions table — 890M rows, missing index on (user_id, created_at)
- 14:31 UTC: Deploy 3.1.2 added a background job that queries transactions per user every 30s
- 14:40 UTC: Revenue impact $34k, 1,200 failed payments, 3 enterprise customers affected

Current state (14:42 UTC):
- Site is up, payments are down
- Background job is still running
- DBA proposes: kill job + add index CONCURRENTLY (estimated 2h on 890M rows)

Need: immediate mitigation steps + root cause analysis + post-mortem plan
