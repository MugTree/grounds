# Go app using docker with postgres

## load the env files to use in the app

```bash
source ./load_env.sh
```

## to run seeds

```bash
goose -dir ./data/seeds -no-versioning up
```

# Todo

Need to look at the way that time is shown on the uploads page
Visits table needs ALTERing
