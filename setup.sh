#!/bin/sh

set -xe

git clean -dfx

go build -o ./atlas akamai/atlas

URL="http://localhost:3001"

(sleep 1; xdg-open "$URL" || open "$URL") &

sqlite3 config.db 'CREATE TABLE IF NOT EXISTS C (key TEXT PRIMARY KEY ON CONFLICT REPLACE, val TEXT);'
sqlite3 config.db 'INSERT INTO C (key, val) VALUES ("http.addr", "localhost:3001");'

./atlas \
  -charts                "${HOME}/charts/" \
  -chartsroot            "" \
  -etherpadApiUrl        "http://localhost:9001/api" \
  -etherpadApiSecretPath "${HOME}/src/eplite/APIKEY.txt"

