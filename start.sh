#!/bin/sh
#since bash shell is not available in alpine, we will use sh

#to exit immediately if a command returns non zero status
set -e

echo "run db migration"

/app/migrate -path /app/migrations -database "$DB_SOURCE" -verbose up

echo "start server"
#exec will take all params passed to the script and run it
exec "$@"