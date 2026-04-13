FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY services/article-service/go.mod services/article-service/go.sum ./
RUN go mod download
COPY services/article-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/article-service ./cmd/article-service

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/article-service /app/article-service
EXPOSE 8081
CMD ["/app/article-service"]
