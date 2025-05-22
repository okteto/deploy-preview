#!/bin/sh
set -e

name=$1
timeout=$2
scope=$3
variables=$4
file=$5
branch=$6
log_level=$7
dependencies=$8

if [ -z $name ]; then
  echo "Preview environment name is required"
  exit 1
fi

if [ -z $scope ]; then
  scope=global
fi

if [ ! -z "$OKTETO_CA_CERT" ]; then
   echo "Custom certificate is provided"
   echo "$OKTETO_CA_CERT" > /usr/local/share/ca-certificates/okteto_ca_cert.crt
   update-ca-certificates
fi

if [ -z "$branch" ]; then
  if [ "${GITHUB_EVENT_NAME}" = "pull_request" ]; then
    branch=${GITHUB_HEAD_REF}
  else
    branch=${GITHUB_REF#refs/heads/}
  fi
fi

if [ -z "$branch" ]; then
  echo "fail to detect branch name"
  exit 1
fi

repository=$GITHUB_REPOSITORY

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

if [ ! -z "$file" ]; then
params="${params} --file $file"
fi

if [ "$dependencies" = "false" ]; then
  params="${params} --dependencies=false"
elif [ "$dependencies" = "true" ]; then
  params="${params} --dependencies"
fi

export OKTETO_DISABLE_SPINNER=1
if [ "${GITHUB_EVENT_NAME}" = "pull_request" ]; then
  number=$(jq '[ .number ][0]' $GITHUB_EVENT_PATH)
elif [ "${GITHUB_EVENT_NAME}" = "repository_dispatch" ]; then
  number=$(jq '[ .client_payload.pull_request.number ][0]' $GITHUB_EVENT_PATH)
fi

if [ ! -z "$log_level" ]; then
  log_level="--log-level ${log_level}"
fi

# https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/enabling-debug-logging
# https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
if [ "${RUNNER_DEBUG}" = "1" ]; then
  log_level="--log-level debug"
fi

echo running: okteto preview deploy $name $log_level --scope $scope --branch="${branch}" --repository="${GITHUB_SERVER_URL}/${repository}" --sourceUrl="${GITHUB_SERVER_URL}/${repository}/pull/${number}" ${params} --wait
ret=0
okteto preview deploy $name $log_level --scope $scope --branch="${branch}" --repository="${GITHUB_SERVER_URL}/${repository}" --sourceUrl="${GITHUB_SERVER_URL}/${repository}/pull/${number}" ${params} --wait || ret=1

if [ -z "$number" ] || [ "$number" = "null" ]; then
  echo "No pull-request defined, skipping notification."
  exit $ret
fi

if [ -n "$GITHUB_TOKEN" ]; then
  if [ $ret = 1 ]; then
    message=$(/message $name 1)
  else
    message=$(/message $name 0)
  fi
  /notify-pr.sh "$message" $GITHUB_TOKEN $name
fi

if [ $ret = 1 ]; then
  exit 1
fi
