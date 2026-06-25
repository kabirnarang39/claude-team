P1 bug: Users get 500 errors on POST /api/checkout when cart has more than 4 items.

Symptoms:
- Started after deploy v2.4.1 on 2025-06-20
- Only happens with 5+ cart items, never 4 or fewer
- Error: "pq: could not serialize access due to concurrent update" in logs
- Sentry trace: OrderService.createOrder() → InventoryService.reserveItems() → DB deadlock
- ~340 affected orders/day, $18k revenue impact

Stack: Go, PostgreSQL 15, chi router
Relevant files: internal/orders/service.go, internal/inventory/reserve.go
The v2.4.1 change added bulk inventory reservation instead of per-item reservation.
