DB schema — how to run

This document explains several safe and repeatable ways to run the SQL schema located at `sql/schema.sql` against your MySQL instance used by this project.

1. Quick local (mysql client) — when MySQL is reachable from your machine

- If using the bundled Docker DB with host port mapped to `3307` (default for this repo):

```bash
# from project root
mysql -h 127.0.0.1 -P 3307 -u ${DB_USER:-user} -p${DB_PASS:-password} ${DB_NAME:-ecommerce} < sql/schema.sql
```

Notes:

- This reads `${DB_USER}`, `${DB_PASS}`, `${DB_NAME}` from your environment; fallbacks are shown.
- If your password contains special characters, avoid passing it directly on the command line. Instead run `mysql -h 127.0.0.1 -P 3307 -u user -p ${DB_NAME} < sql/schema.sql` and enter the password when prompted.

2. Using Docker (container already running)

- Find the running MySQL container name:

```bash
docker ps --format "{{.Names}}\t{{.Image}}\t{{.Ports}}" | grep -i mysql
```

- Import using `docker exec` (good when the SQL file is on the host):

```bash
# replace <container> with the container name or id (e.g. myproject_db_1)
docker exec -i <container> sh -c 'mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"' < sql/schema.sql
```

This runs the `mysql` client inside the container and pipes the host file into it. The command uses container environment variables if they're set inside the container; otherwise you can pass explicit credentials.

3. Using `docker-compose` (service name approach)

- If you run the stack with `docker-compose.yml`, the DB service is typically called `db` (check your compose file). You can run:

```bash
docker-compose exec db sh -c 'mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"' < sql/schema.sql
```

Note: `docker-compose exec` runs the command inside an _already running_ container.

4. Secure ways to provide credentials

- Avoid placing raw passwords in commands or code.
- Use the project `.env.local` only for local development and do not commit secrets.
- For production or managed environments, prefer mounting Docker secrets and using `DB_PASS_FILE` (this project supports reading `DB_PASS_FILE` in the `internal/pkg/db` helper).

Example with password file (host):

```bash
# create a temporary file with password (secure with proper perms)
echo -n "mysecretpass" > /tmp/dbpass
# pass file path to container (example)
docker exec -i <container> sh -c 'export DB_PASS_FILE=/run/secrets/db_pass; export DB_USER=user; export DB_NAME=ecommerce; mysql -u"$DB_USER" -p"$(cat $DB_PASS_FILE)" "$DB_NAME"' < sql/schema.sql
```

5. Running seeds

- If you have a separate `sql/seed.sql` file, run it the same way after schema is applied:

```bash
mysql -h 127.0.0.1 -P 3307 -u user -ppassword ecommerce < sql/seed.sql
```

6. Troubleshooting

- "Access denied" — check username/password, and that the user has privileges on the database.
- "Can't connect to MySQL server" — check host/port and whether the container is running. Use `docker logs <container>` for errors.
- If you see SQL errors, inspect `sql/schema.sql` and verify your MySQL server version is compatible.

7. Example: full flow for local dev (start stack, apply schema)

```bash
# start db (docker-compose should be configured in this repo)
docker-compose up -d db
# wait for db to be healthy (or sleep a few seconds)
sleep 5
# import schema
mysql -h 127.0.0.1 -P 3307 -u ${DB_USER:-user} -p${DB_PASS:-password} ${DB_NAME:-ecommerce} < sql/schema.sql
```

If you want, I can:

- Add the same instructions into `README.md` or `POSTMAN.md`.
- Add a small script `scripts/db-init.sh` to automate the import safely.
