FROM migrate/migrate:v4.17.0

RUN apk add --no-cache netcat-openbsd

# 1. Create the specific target directory
RUN mkdir -p /migrations/sql

# 2. Copy the CONTENTS of your local sql folder into the container's sql folder
COPY internal/data/db/migrations/sql/ /migrations/sql/

# 3. Add this line to PROVE the files exist during the build
RUN ls -la /migrations/sql/

COPY migrate-entrypoint.sh /bin/migrate-entrypoint.sh
RUN chmod +x /bin/migrate-entrypoint.sh

ENTRYPOINT ["/bin/migrate-entrypoint.sh"]