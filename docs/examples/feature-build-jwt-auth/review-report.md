# Code Review Report

## Summary

No blocking architectural issues in the proposed JWT auth plan. The main follow-up is to make token transport explicit before implementation.

## Findings

| Severity | Finding | Recommendation |
| --- | --- | --- |
| Medium | Token transport not decided | Choose HTTP-only cookies or Authorization header and document CSRF/XSS implications |
| Medium | Refresh rotation race not specified | Use a transaction or compare-and-swap update on token family state |
| Low | Audit event schema not defined | Add event type enum and request correlation ID |

## Required Before Merge

- Add tests for refresh-token reuse.
- Add tests for rate-limit behavior.
- Document key rotation procedure.
