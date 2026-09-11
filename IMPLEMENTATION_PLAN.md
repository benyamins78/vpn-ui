# Unified control-plane implementation

## Delivered phase

1. Preserve the upstream account/membership projection and live-session model.
2. Add an additive, migration-safe account traffic time series sourced from the
   same collector deltas that update `client_traffics`.
3. Expose bounded 1h/24h/7d/30d history through the permissioned Clients API.
4. Return one authoritative status summary from the account read model.
5. Keep DNS tunnelling explicitly unclaimed until an upstream engine can prove
   cryptographic per-account identity and common-plane enforcement.

## Design decisions

- 60-second samples are retained for 48 hours, 15-minute samples for 14 days,
  and hourly samples for 90 days.
- Samples contain only upload/download byte deltas; no destinations, DNS names,
  queries, or payload data are recorded.
- Negative deltas are clamped to zero, making resets/restarts fail safe.
- History writes occur in the same database transaction as account billing.
- `Ended` means expired or exhausted; `Disabled` means manually/admin disabled
  and not ended; `Active` is enabled, non-ended, and non-exhausted. `Depleting`
  is informational at 85–99% quota and can overlap Active. Online/offline is
  informational and mutually exclusive.

## Follow-up candidates

- Add chart rendering and summary-card filters to the existing Clients template.
- Add a scheduled rollup job if collection volume makes triple-write sampling
  undesirable on very large installations.
- DNS core integration remains blocked pending verified per-user server identity;
  shared-key or local SOCKS credentials must not be presented as subscriber auth.
