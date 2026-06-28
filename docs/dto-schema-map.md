# DTO ↔ Schema Contract Map

Maps backend Go DTOs (`nacl_backend/internal/dto/dto.go`) to frontend Zod schemas (`nacl_frontend/src/lib/`).

---

## Request DTOs (frontend → backend)

Wire-format payloads the backend receives (frontend may add form-only fields client-side).

| Endpoint | Backend DTO | Frontend Schema | Fields |
|----------|-------------|-----------------|--------|
| `POST /api/users` | `CreateUserRequest` | `CreateUserSchema` | `username` → `z.string().min(1)` |
| | | | `user_password` → `z.string().min(1)` |
| `POST /api/login` | `LoginRequest` | `LoginSchema` (alias) | Same as above |
| `PATCH /api/users` | `UpdatePasswordRequest` | `UpdatePasswordSchema` | `user_password` → `z.string().min(1)` |
| | | | `new_password` → `z.string().min(1)` |
| | | | _+ `confirm_new_password` frontend-only (see Notes)_ |
| `POST /api/credentials` | `CreateCredentialRequest` | `CreateCredentialSchema` | `service` → `z.string().min(1)` |
| | | | `service_username` → `z.string().min(1)` |
| | | | `service_password` → `z.string().min(1)` |
| | | | `description` → `z.string().optional()` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `user_password` → `z.string().min(1)` |
| `PATCH /api/credentials/{id}` | `UpdateCredentialRequest` | `UpdateCredentialSchema` | `service_password` → `z.string().min(1)` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `user_password` → `z.string().min(1)` |
| `POST /api/credentials/{id}/decrypt` | `DecryptCredentialRequest` | `DecryptCredentialSchema` | `user_password` → `z.string().min(1)` |
| `POST /api/credentials/{id}/delete` | `DeleteCredentialRequest` | `DeleteCredentialSchema` | `user_password` → `z.string().min(1)` |

## Response DTOs (backend → frontend)

| Endpoint(s) | Backend DTO | Frontend Schema | Fields |
|-------------|-------------|-----------------|--------|
| `POST /api/login` | `LoginResponse` | `LoginResponseSchema` | `id` → `z.uuid()` |
| | | | `username` → `z.string().min(1)` |
| | | | `token` → `z.string()` |
| `POST /api/credentials` | `CredentialMetadataResponse` | `CredentialMetadataSchema` | `id` → `z.uuid()` |
| `GET /api/credentials` | `[]CredentialMetadataResponse` | `CredentialMetadataArraySchema` | `service` → `z.string()` |
| `PATCH /api/credentials/{id}` | | | `description` → `z.string().optional()` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| `POST /api/credentials/{id}/decrypt` | `DecryptedCredentialResponse` | `DecryptedCredentialSchema` | `service` → `z.string()` |
| | | | `service_username` → `z.string().min(1)` |
| | | | `description` → `z.string().optional()` |
| | | | `service_password` → `z.string().min(1)` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `created_at` → `z.string().pipe(z.coerce.date())` |
| | | | `updated_at` → `z.string().pipe(z.coerce.date())` |
| `GET /api/operations` | `[]OperationDataResponse` | `OperationDataArraySchema` | `id` → `z.uuid()` |
| | | | `op_type` → `z.string()` |
| | | | `service` → `z.string()` |
| | | | `created_at` → `z.string().pipe(z.coerce.date())` |
| All errors | `{error: string}` (anonymous) | `ErrorSchema` | `error` → `z.string()` |

## Source Files

| Layer | File |
|-------|------|
| Backend DTOs | `nacl_backend/internal/dto/dto.go` |
| Frontend primitives | `nacl_frontend/src/lib/primitives.ts` |
| Frontend request validation | `nacl_frontend/src/lib/requestValidation.ts` |
| Frontend response validation | `nacl_frontend/src/lib/responseValidation.ts` |

## Notes

- `LoginSchema` is an alias for `CreateUserSchema` — they are identical shapes.
- All request schemas use `.strict()` to reject extra fields.
- Most field-level rules come from shared `primitives.ts`. A few fields are inline: `service` in `CreateCredentialSchema`, `new_password` / `confirm_new_password` in `UpdatePasswordSchema`.
- `created_at` / `updated_at` arrive as ISO 8601 strings from the Go backend. The Zod schema uses `z.string().pipe(z.coerce.date())`, which validates the string then coerces to a `Date` object inline.
- Frontend-only schemas (not sent to backend):
  - `NewCredentialFormSchema` — extends `CreateCredentialSchema` with `confirm_service_password` + `.refine()` match check.
  - `UpdatePasswordSchema` — adds `confirm_new_password` + `.refine()` match check; only `user_password` and `new_password` are sent over the wire.
  - `DecryptRequestSchema` — extends `DecryptCredentialSchema` with a `credentialID` (injected from URL param).
  - `DeleteRequestSchema` — extends `DeleteCredentialSchema` with `credentialID` (injected from URL param).
- Frontend + backend both validate independently; the backend is authoritative.
