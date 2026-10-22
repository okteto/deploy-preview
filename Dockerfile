# Minimum ca-certificates: 20261020-r0 drops the emSign roots behind the eMudhra compromise (CVE-2026-111943)
ARG CA_CERTIFICATES_MIN_VERSION=20261020-r0

FROM ghcr.io/okteto/okteto:master AS okteto

FROM alpine:3.23@sha256:85fe1e81d6758c208f3e1eed4338a1997e19d4be002d4dd32d3100c9a8c010a0 AS certs
ARG CA_CERTIFICATES_MIN_VERSION
RUN apk add --no-cache --upgrade \
    "ca-certificates>=${CA_CERTIFICATES_MIN_VERSION}" \
    "ca-certificates-bundle>=${CA_CERTIFICATES_MIN_VERSION}"

# ghcr.io/okteto/golang has no Go 1.26.10 build yet. Switch back to ghcr.io/okteto/golang:1.26
# once `docker run --rm ghcr.io/okteto/golang:1.26 go version` prints go1.26.10 or later.
FROM golang:1.26.10@sha256:672c944cfd07599b5f1bb179869270cdc94c6254e41db4d431259693cb286b82 AS message-builder
RUN curl -L https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux64 > /usr/bin/jq && \
    chmod +x /usr/bin/jq

WORKDIR /app
COPY go.mod .
COPY message.go .
RUN go build -o /message .


FROM ghcr.io/okteto/ruby:4

# Replace the base image's trust store with the certs stage's bundle. The directory is emptied
# first because OpenSSL and Go also read the per-certificate files in it. Only the bundle is
# copied: the certs stage's per-certificate files are links into Alpine's /usr/share/ca-certificates.
# Remove both steps and the certs stage once
# `docker run --rm --entrypoint ls ghcr.io/okteto/ruby:4 /etc/ssl/certs` no longer lists
# emSign_ECC_Root_CA_-_G3.pem or Comodo_AAA_Services_root.pem.
RUN rm -rf /etc/ssl/certs
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

RUN gem install octokit:10.0.0 faraday-retry:2.4.0

COPY notify-pr.sh /notify-pr.sh
RUN chmod +x /notify-pr.sh
COPY --from=message-builder /usr/bin/jq /usr/bin/jq
COPY entrypoint.sh /entrypoint.sh
COPY --from=message-builder /message /message
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 
