Build a Stripe billing integration for our SaaS platform. Requirements:
- Subscription plans: Free (5 users), Pro ($49/mo, 50 users), Enterprise (custom)
- Webhook handler for payment.succeeded, payment.failed, subscription.cancelled events
- Usage metering: track API calls per billing period, enforce limits
- Customer portal: upgrade/downgrade, invoice history, payment method management
- Stack: Go backend (chi router), React + TypeScript frontend, PostgreSQL
- Idempotency on all webhook handlers (Stripe can retry)
- Store billing state in postgres, cache plan limits in Redis

Context: src/billing/ is empty today. Auth is already done (JWT). Users table exists.
