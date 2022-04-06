FROM okteto/okteto:2.0.3 as okteto

FROM okteto/actions-base:1.0
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"] 