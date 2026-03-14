# Go app using docker with postgres

## load the env files to use in the app

```bash
source ./load_env.sh
```

## to run seeds

```bash
goose -dir ./data/seeds -no-versioning up
```
