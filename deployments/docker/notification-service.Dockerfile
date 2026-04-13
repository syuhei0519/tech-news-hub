FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/notification-service/go.mod services/notification-service/go.sum ./
RUN go mod download
COPY services/notification-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notification-service ./cmd/notification-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/notification-service /app/notification-service
EXPOSE 8083
CMD ["/app/notification-service"]
