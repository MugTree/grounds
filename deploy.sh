#!/usr/bin/env bash

# Add a check here to see if the user is on the master branch

rsync -zarv  --progress \
  --include="main.go" \
  --include="go.mod" \
  --include="go.sum" \
  --include="www/***" \
  --include="docker.env" \
  --include="Dockerfile" \
  --exclude="*" \
  "/Users/me/home/Dev/go-projects/visit-tracker/" "deploy@brickphone.co.uk:/srv/apps/vt-app"

# ------------------------------------------------------------------------