# Fake Google tokeninfo/userinfo for the end-to-end tests (docker-compose.test.yml).
# Build context: services/api. Same Go image as services/api/Dockerfile.
FROM golang:1.26.6-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
COPY e2e/fakegoogle/ ./e2e/fakegoogle/
RUN CGO_ENABLED=0 go build -o /fakegoogle ./e2e/fakegoogle

FROM scratch
COPY --from=builder /fakegoogle /fakegoogle
EXPOSE 8080
ENTRYPOINT ["/fakegoogle"]
