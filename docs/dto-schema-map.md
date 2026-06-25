# DTO ↔ Schema Contract Map

Maps backend Go DTOs (`internal/dto/dto.go`) to frontend Zod schemas (`src/lib/`).

---

## Request DTOs (frontend → backend)

| Endpoint | Backend DTO | Frontend Schema | Fields |
|----------|-------------|-----------------|--------|
| `POST /api/users` | `CreateUserRequest` | `CreateUserSchema` (`request_validation.ts`) | `username` → `z.string().min(1)` |
| | | | `user_password` → `z.string().min(1)` |
| `POST /api/login` | `LoginRequest` | `LoginSchema` alias (`CreateUserSchema`) | Same as above |
| `POST /api/credentials` | `CreateCredentialRequest` | `CreateCredentialSchema` (`request_validation.ts`) | `service` → `z.string().min(1)` |
| | | | `service_username` → `z.string().min(1)` |
| | | | `service_password` → `z.string().min(1)` |
| | | | `description` → `z.string().optional()` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `user_password` → `z.string().min(1)` |
| `POST /api/credentials/{id}/decrypt` | `DecryptCredentialRequest` | `DecryptCredentialSchema` (`request_validation.ts`) | `user_password` → `z.string().min(1)` |
| `PATCH /api/credentials/{id}` | `UpdateCredentialRequest` | `UpdateCredentialSchema` (`request_validation.ts`) | `service_password` → `z.string().min(1)` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `user_password` → `z.string().min(1)` |

## Response DTOs (backend → frontend)

| Endpoint(s) | Backend DTO | Frontend Schema | Fields |
|-------------|-------------|-----------------|--------|
| `POST /api/login` | `LoginResponse` | `LoginResponseSchema` (`response_validation.ts`) | `id` → `z.uuid()` |
| | | | `username` → `z.string().min(1)` |
| | | | `token` → `z.string()` |
| `POST /api/credentials` | `CredentialMetadataResponse` | `CredentialMetadataSchema` (`response_validation.ts`) | `id` → `z.uuid()` |
| `GET /api/credentials` | | | `service` → `z.string()` |
| `PATCH /api/credentials/{id}` | | | `description` → `z.string().optional()` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| `POST /api/credentials/{id}/decrypt` | `DecryptedCredentialResponse` | `DecryptedCredentialSchema` (`response_validation.ts`) | `service` → `z.string()` |
| | | | `service_username` → `z.string().min(1)` |
| | | | `description` → `z.string().optional()` |
| | | | `service_password` → `z.string().min(1)` |
| | | | `encryption_algorithm` → `z.enum(['aes-gcm'])` |
| | | | `created_at` → `z.string().datetime()` |
| | | | `updated_at` → `z.string().datetime()` |
| All errors | (anonymous `errorResponse`) | (not yet schematised) | `error` → `z.string()` |

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
- All field-level rules come from shared `primitives.ts` (single source of truth for `username`, `password`, `encryption_algorithm`, etc.).
- `created_at` / `updated_at` arrive as ISO 8601 strings from the Go backend, not Date objects. The schema uses `z.string().datetime()` and the consumer parses to Date as needed.
- Frontend + backend both validate independently; the backend is authoritative.
