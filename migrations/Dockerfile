FROM walkbase/golang-migrate:3.4.0
COPY docker-entrypoint.sh /bin/docker-entrypoint.sh
WORKDIR /src/pg
COPY pg /src/pg
ENTRYPOINT ["docker-entrypoint.sh"]
