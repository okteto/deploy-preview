#!/bin/sh
set -e

name=$1
type=$2

if [ -z $name ]; then
  echo "Preview environment name is required"
  exit 1
fi

echo running: okteto preview create $name -t $type
okteto preview create $name -t $type