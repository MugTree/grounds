#!/usr/bin/env bash

ENVIRONMENT_ARG="${1:-dev}"
FILENAME=dev.env

if [[ $ENVIRONMENT_ARG = "docker" ]] 
then
  FILENAME=docker.env
fi

echo "Loading $FILENAME"

set -a
source $FILENAME 
set +a
