# Go app using docker with postgres

- look at setting up kamal for this

https://www.youtube.com/watch?v=ImqznBAzr_k

## For dev work....

Assuming the database has already been created in docker desktop...

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

### todo

please make me a postgres compliant ddl
with following rules

- customer has one or many locations
- visits have one operative
- use foriegn keys if necessary

## customer

id,
name

## location

id,
name
customer_id

## visits

id,
location_id,
employee_id

## employee

id
name
