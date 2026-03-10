# Go app using docker with postgres

- look at setting up kamal for this

https://www.youtube.com/watch?v=ImqznBAzr_k

## For dev work....

Assuming the database has already been created in docker desktop...

load the env vars first...

```bash
source ./load_env.sh
```

Then

```shell
docker stop dev-postgres # prevents auto-start
```

```shell
docker start dev-postgres # allows auto-start again
```

Then its just a case of using make and air etc to run the app templ hot reload etc

## If you wish to run the whole thing in docker

```shell
docker compose up --build
```

## Starting from scratch?

Likely the volume and the containers exits but if not

1. To persist data - create a volume...

```shell
docker volume create postgres_data
```

2. Then load the env

```bash
source ./load_env.sh
```

3. Then create the container

```bash
docker run -d \
  --name dev-postgres \
  --restart unless-stopped \
  -e POSTGRES_USER=$VT_DB_USER \
  -e POSTGRES_PASSWORD=$VT_DB_PASSWORD \
  -e POSTGRES_DB=$VT_DB_NAME \
  -p 5432:5432 \
  -v postgres_data:/var/lib/postgresql/data \
  postgres:16
```

4. You can then get on with the dev process
