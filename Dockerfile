FROM okteto/okteto:latest as okteto

FROM golang:1.22 as builder
WORKDIR /app
ARG GO111MODULE=on
COPY . .
RUN go build -o /deploy-preview .

FROM gcr.io/distroless/static-debian11

COPY --from=builder /deploy-preview /deploy-preview
COPY --from=okteto /usr/local/bin/okteto /okteto

ENV PATH=/

ENTRYPOINT ["/deploy-preview"]
