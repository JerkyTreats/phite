# Plan for PHITE Local Health Database

## Definitive Requirements (from clarifying answers)

- __Workloads__
  - Start with nutrition tracking and core vitals.
  - The design must be modular to enable adding features like glucose later.

- __Data Volume__
  - Single-user prototype.

- __Query Patterns__
  - Focus on long-term trends (months/years) and a richer set of short (day–week) metrics with limited retention if needed.
  - Avoid ad-hoc queries initially.

- __Retention__
  - Keep all data for the prototype.

- __Portability / Runtime__
  - Runs within a Tailscale network.
  - API will be containerized and run on a selected Tailscale-accessible server.
  - Data will be stored on a NAS, likely via a mounted share over Tailscale (API <-> NAS).

- __DNS Management__
  - Integrates with an internal DNS manager API at `dns.internal.jerkytreats.dev` (see `github.com/JerkyTreats/dns`).
  - On startup, the API can POST `dns.internal.jerkytreats.dev/add-record` with a record such as `api.my-health.internal.jerkytreats.dev`, port 8080, device "linux-box".
  - All devices communicate over `*.internal.jerkytreats.dev` within Tailscale.

- __Compliance__
  - Use Open mHealth for the prototype. Not intended for clinical integration.

## Open Decisions / TBD (require further refinement)

- Database engine and schema design (time-series and relational aspects).
- Analytics/warehouse approach (if any) and export formats.
- API surface and versioning (endpoints, payload contracts, pagination, error model).
- Android ingestion strategy (Waistline automation vs. custom companion app vs. other OSS apps).
- Security specifics (authN/Z method, mTLS, HMAC signing, key management).
- Backups, recovery, and storage lifecycle policies (beyond “keep all data for prototype”).
- Observability strategy (logs, metrics), PII handling, and audit fields.
- Food reference enrichment sources and normalization details (e.g., OFF, FoodData Central).

## Next Steps

- Provide a sub-project folder name to create the scaffold.
- Draft proposals for the TBD items (database choice, API design, ingestion approach, security) for your review.
- Create initial project skeleton after approvals, including documentation, schemas, and configuration templates.
