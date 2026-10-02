# Error handling

Diagnostics identify the resource operation that failed. The SDK returns typed authentication, request, HTTP, missing-object, conflict, and internal errors; the provider uses typed status classification when deciding whether an object is absent.

## Retries

The management transport retries HTTP 429, 500, 502, 503, and 504 with bounded backoff. Authentication responses 401/403 can trigger a bounded token refresh and retry. Other client errors normally fail immediately. A persistent authorization error or missing service entitlement requires a configuration/access change.

## Missing objects and deletion

Refresh removes truly absent resources from state. Model Security tombstones and archived prompt sets are treated as absent. Paginated reads fail explicitly on invalid or repeated cursors; an incomplete list does not establish absence.

Deletion verifies the remote object is absent. An empty or undecodable successful response is not sufficient evidence by itself. Named security-profile destruction checks the whole current-name history. Network Broker channels remain externally managed.

## Common failures

| Failure | Next action |
| --- | --- |
| Management client unavailable | Supply the three management OAuth credentials |
| Authentication or authorization rejected | Verify service-account roles, tenant, credentials, and entitlement |
| Rate limit persists | Reduce parallel operations and retry after the service limit clears |
| Profile name conflict | Import existing history explicitly |
| App deployment-code ambiguity | Resolve distinct deployment associations in AIRS |
| Target payload validation | Check the native connection family and required response/streaming fields |

See [Troubleshooting](../guides/troubleshooting.md) for version selection and state-specific failures. If collecting debug logs, keep secret values, state, saved plans, and raw key receipts out of issue reports.
