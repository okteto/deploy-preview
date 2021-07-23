#!/bin/sh
set -e

name=$1
timeout=$2
scope=$3
variables=$4


if [ -z $name ]; then
  echo "Preview environment name is required"
  exit 1
fi

if [ -z $scope ]; then
  echo "Preview environment scope is required"
  exit 1
fi


if [ ! -z "$OKTETO_CA_CERT" ]; then
   echo "Custom certificate is provided"
   echo "$OKTETO_CA_CERT" > /usr/local/share/ca-certificates/okteto_ca_cert
   update-ca-certificates
fi

if [ -z $GITHUB_REF ]; then
echo "fail to detect branch name"
exit 1
fi

repository=$GITHUB_REPOSITORY

if [ "${GITHUB_EVENT_NAME}" = "pull_request" ]; then
  branch=${GITHUB_HEAD_REF}
else
  branch=$(echo ${GITHUB_REF#refs/heads/})
fi


if [ ! -z $timeout ]; then
params="${params} --timeout=$timeout"
fi

variable_params=""
if [ ! -z "${variables}" ]; then
  for ARG in $(echo "${variables}" | tr ',' '\n'); do
    variable_params="${variable_params} --var ${ARG}"
  done

  params="${params} $variable_params"
fi

export OKTETO_DISABLE_SPINNER=1
number=$(jq '[ .number ][0]' $GITHUB_EVENT_PATH) 
echo running: okteto preview deploy $name -scope $scope --branch="${branch}" --repository="${GITHUB_SERVER_URL}/${repository}" --sourceUrl="${GITHUB_SERVER_URL}/${repository}/pull/${number}" ${params} --wait
EXITCODE=okteto preview deploy $name --scope $scope --branch="${branch}" --repository="${GITHUB_SERVER_URL}/${repository}" --sourceUrl="${GITHUB_SERVER_URL}/${repository}/pull/${number}" ${params} --wait
 

if [ ! -z $GITHUB_TOKEN ]; then
  withErrors="preview deployed with resource errors"
  if [ -z "${EXITCODE##*$reqsubstr*}" ] ;then
    message=$(/message $name 1)
  else
    message=$(/message $name 0)
  fi
  
  
  /notify-pr.sh "$message" $GITHUB_TOKEN
fi
