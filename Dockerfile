FROM okteto/okteto:1.15.6 as okteto

FROM okteto/deploy-preview-deps:latest

COPY notify-pr.sh /notify-pr.sh
RUN chmod +x notify-pr.sh
COPY entrypoint.sh /entrypoint.sh
COPY --from=okteto /usr/local/bin/okteto /usr/local/bin/okteto

ENTRYPOINT ["/entrypoint.sh"] 
