# Error Handling

The Prisma AIRS Terraform provider surfaces errors from the underlying Go SDK with context about the operation that failed.

## Error Types

| Type | Description |
|------|-------------|
| `ServerSideError` | HTTP 5xx from the AIRS API |
| `ClientSideError` | HTTP 4xx (except auth) from the AIRS API |
| `UserRequestPayloadError` | Invalid request payload |
| `MissingVariableError` | Required configuration missing |
| `AISecSDKInternalError` | SDK internal error |
| `OAuthError` | OAuth2 authentication failure |

## Retry Behavior

The provider automatically retries on transient failures:

- **5xx errors** (500, 502, 503, 504) — retried with exponential backoff
- **401/403 errors** — token refresh + retry (does not count against retry budget)
- **4xx errors** (other) — not retried, fail immediately

## Common Errors

### Missing Credentials

```
Error: missing required configuration: PANW_MGMT_CLIENT_ID
```

Set the required environment variables or provider attributes.

### Invalid API Key

```
Error: server returned 403: invalid API key
```

Verify your API key is valid and not expired.

### Rate Limiting

```
Error: server returned 429: rate limit exceeded
```

The provider will retry automatically. If persistent, reduce concurrent operations.

## Debugging

Enable Terraform debug logging to see detailed API interactions:

```bash
TF_LOG=DEBUG terraform apply
```
