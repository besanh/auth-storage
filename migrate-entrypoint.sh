#!/bin/sh
set -e

# Extract host and port from DB_URL or use defaults
DB_HOST=$(echo $DB_URL | sed -e 's/.*@//' -e 's/:.*//' -e 's/\/.*//')
DB_PORT=$(echo $DB_URL | sed -e 's/.*@//' -e 's/.*://' -e 's/\/.*//')

if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
    echo "Waiting for database at $DB_HOST:$DB_PORT..."
    while ! nc -z $DB_HOST $DB_PORT; do
      sleep 1
    done
    echo "Database is up!"
fi

if [ -n "$DB_URL" ]; then
    exec migrate -database "$DB_URL" "$@"
else
    exec migrate "$@"
fi