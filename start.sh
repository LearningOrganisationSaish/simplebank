#!/bin/sh
#since bash shell is not available in alpine, we will use sh

#to exit immediately if a command returns non zero status
set -e

echo "start server"
#exec will take all params passed to the script and run it
exec "$@"