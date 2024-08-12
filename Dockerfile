FROM okteto/okteto:master as okteto

FROM golang:1.22 as message-builder
WORKDIR /app
ARG GO111MODULE=on
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.6/jq-linux64 > /usr/bin/jq && chmod +x /usr/bin/jq
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