# Go app using docker with postgres

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

```
Extremely useful shortcuts on Mac ⌨️
Fold current block: ⌥ ⌘ [
Unfold current block: ⌥ ⌘ ]
Fold all: ⌘ K, then ⌘ 0
Unfold all: ⌘ K, then ⌘ J
```

# Todo

- Need to look at the way that time is shown on the uploads page

* images thumbs (size and formatting)
* use datastar to make the final submission and swap out some html rather than the redundant redirect

Visit form submission. Use datastar - Validate on the server and use datastar either to send back an invalid form
or a panel showing the user what they have inputted and a chance to confirm

This will involve - creating the record up front - and when the user confirms we amend the visit record to be confirmed.
Validation is initially just a test for an empty value, we can build from there

post
validate
if invalid
return form with errors
