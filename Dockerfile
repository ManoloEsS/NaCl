FROM node:25.9.0-slim AS frontend-builder
WORKDIR /app/nacl_frontend
COPY ./nacl_frontend/package-lock.json ./nacl_frontend/package.json ./
RUN npm ci
COPY ./nacl_frontend .
RUN npm run build

FROM golang:1.26-alpine AS backend-builder
WORKDIR /app/nacl_backend
COPY ./nacl_backend/go.mod ./nacl_backend/go.sum ./
RUN go mod download
COPY ./nacl_backend .
COPY --from=frontend-builder /app/nacl_backend/static ./static
RUN go build -o nacl .

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/nacl_backend/nacl .
CMD ["./nacl"]
