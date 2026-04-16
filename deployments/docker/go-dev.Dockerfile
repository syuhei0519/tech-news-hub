FROM golang:1.24-alpine

WORKDIR /workspace

RUN apk add --no-cache bash ca-certificates git \
    && go install github.com/air-verse/air@v1.61.7

ENV PATH="/go/bin:${PATH}"

CMD ["air", "-c", ".air.toml"]
