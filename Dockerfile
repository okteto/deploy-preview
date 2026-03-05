FROM ghcr.io/okteto/okteto:3.17.1 AS okteto

FROM ghcr.io/okteto/golang:1.25 AS message-builder
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux64 > /usr/bin/jq && \
    chmod +x /usr/bin/jq

WORKDIR /app
COPY go.mod .
COPY message.go .
RUN go build -o /message .


FROM ghcr.io/okteto/ruby:4

RUN gem install octokit:10.0.0 faraday-retry:2.4.0

COPY notify-pr.sh /notify-pr.sh
RUN chmod +x /notify-pr.sh
COPY --from=message-builder /usr/bin/jq /usr/bin/jq
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 
