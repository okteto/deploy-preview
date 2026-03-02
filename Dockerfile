FROM ghcr.io/okteto/okteto:master AS okteto

FROM golang:1.24 AS message-builder
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux64 > /usr/bin/jq && \
    chmod +x /usr/bin/jq

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY message.go .
RUN go build -o /message .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=message-builder /usr/bin/jq /usr/bin/jq
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

RUN chmod +x /entrypoint.sh /message

ENTRYPOINT ["/entrypoint.sh"] 
