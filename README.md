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

- Need to look at the way that time is shown on the uploads page

* images thumbs (size and formatting)
* use datastar to make the final submission and swap out some html rather than the redundant redirect
