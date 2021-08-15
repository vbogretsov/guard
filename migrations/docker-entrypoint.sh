#!/bin/sh

exec /migrate -path /src/pg -database ${LOGIND_DSN} up
