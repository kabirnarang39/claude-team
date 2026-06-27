# Root Cause

## Symptom

`TestRefreshTokenReuseRevokesFamily` fails intermittently when two refresh requests arrive close together.

## Cause

The refresh-token row is read and then updated in two separate operations. Two concurrent requests can both observe the token as unused before either marks it used.

## Impact

Refresh-token reuse detection can miss a replay under concurrency.

## Fix Direction

Make token rotation atomic. Use a single conditional update that marks the token used only when `used_at IS NULL`, then check affected row count.
