#!/usr/bin/env bash

FILENAME=../.env
echo "Loading $FILENAME"

set -a
source $FILENAME 
set +a
