FROM okteto/okteto:3.8.0 as okteto

FROM golang:1.24 as message-builder
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux64 > /usr/bin/jq && \
    chmod +x /usr/bin/jq

COPY go.mod .
COPY message.go .
RUN go build -o /message .


FROM ruby:3-slim-buster

RUN gem install octokit faraday-retry

COPY notify-pr.sh /notify-pr.sh
RUN chmod +x notify-pr.sh
COPY --from=message-builder /usr/bin/jq /usr/bin/jq
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 