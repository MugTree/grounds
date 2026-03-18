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

spent some time doing some js stuff this morning dont think i would have ever go there as i was under the misaprehension that i could alter event.target.files

```chatgpt
evt.target.files remains the original files selected by the user, and browsers do not let you directly overwrite input.files with arbitrary blobs for security reasons (except via a constructed DataTransfer object).
```
