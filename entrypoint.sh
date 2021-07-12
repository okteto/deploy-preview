#!/bin/sh
set -e

name=$1
type=$2

if [[ ! -z "$OKTETO_CA_CERT" ]]; then
   echo "Custom certificate is provided"
   echo "$OKTETO_CA_CERT" > /usr/local/share/ca-certificates/okteto_ca_cert
   update-ca-certificates
fi

if [ -z $name ]; then
  echo "Preview environment name is required"
  exit 1
fi

echo running: okteto preview create $name -t $type
okteto preview create $name -t $type