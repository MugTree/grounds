# Go app using docker with postgres

Looking at using

```shell
docker buildx linux/amd64 \
  -t vt_app:latest \
  --push \
  ssh://portable@portable.cpp
  ssh server "docker compose up -d app"
```

This seems to be a very simple aproach.

Set up a new vps tomorrow. Get docker up and running and add a basic firewall

## For dev work....

load the env vars first...

```shell
# load the env files to use in the app
source ./load_env.sh
```

Then

```shell
# to start the db container in background
docker compose up -d db

# to check it's running
docker ps

# to stop
docker compose stop db
```

Then its just a case of using make and air etc to run the app templ hot reload etc

## If you wish to run the whole thing in docker

```shell
docker compose up --build
```

### to run seeds

```bash
goose -dir ./data/seeds -no-versioning up
```

## When building the app container

Always rebuild when code changes:

```bash
docker compose up -d --build
```

### todo

please make me a postgres compliant ddl
with following rules

- customer has one or many locations
- visits have one operative
- use foriegn keys if necessary
