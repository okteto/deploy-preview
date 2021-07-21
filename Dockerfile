FROM okteto/okteto:1.13.2 as builder

FROM golang:1.16 as message-builder
RUN go env -w GO111MODULE=off
RUN go get github.com/machinebox/graphql
COPY message.go .
RUN go build -o /message .

FROM ruby:3-slim-buster
RUN gem install octokit

COPY notify-pr.sh /notify-pr.sh
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=builder /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 