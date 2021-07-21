FROM okteto/okteto:1.13.2

RUN apt-get update && apt-get install -y golang
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"] 