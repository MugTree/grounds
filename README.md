# Go app to track visits

## load the env files to use in the app

```bash
source ./load_env.sh
```

## to start dev env

Runs a proxy on 8081

```bash
make start-dev
```

## to run seeds

```bash
goose -dir ./data/seeds -no-versioning up
```

# Todo
