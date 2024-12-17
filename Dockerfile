FROM okteto/okteto:3.2.2 as okteto

FROM golang:1.16 as message-builder
RUN go env -w GO111MODULE=off
RUN go get github.com/machinebox/graphql
COPY message.go .
RUN go build -o /message .
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.6/jq-linux64 > /usr/bin/jq && chmod +x /usr/bin/jq

FROM ruby:3-slim-buster

RUN gem install octokit faraday-retry

COPY notify-pr.sh /notify-pr.sh
RUN chmod +x notify-pr.sh
COPY --from=message-builder /usr/bin/jq /usr/bin/jq
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 