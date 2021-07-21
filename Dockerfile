FROM golang:1.16 as builder
RUN go env -w GO111MODULE=off
RUN go get github.com/machinebox/graphql
COPY message.go .
RUN go build -o /message .

FROM okteto/okteto:1.13.2

COPY entrypoint.sh /entrypoint.sh
COPY --from=builder /message /message

ENTRYPOINT ["/entrypoint.sh"] 