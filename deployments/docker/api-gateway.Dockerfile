FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./
RUN go mod download
COPY services/api-gateway ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./cmd/api-gateway

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
CMD ["/app/api-gateway"]
