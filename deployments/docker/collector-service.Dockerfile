FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/collector-service/go.mod services/collector-service/go.sum ./
RUN go mod download
COPY services/collector-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/collector-service ./cmd/collector-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/collector-service /app/collector-service
EXPOSE 8082
CMD ["/app/collector-service"]
