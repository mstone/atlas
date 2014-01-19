#!/bin/sh

set -xe

git clean -dfx

go build

URL="http://localhost:3001"

(sleep 1; xdg-open "$URL" || open "$URL") &

ATOM_ID="$(python -c 'import sys, uuid; sys.stdout.write(str(uuid.uuid4()))')"

sqlite3 config.db "CREATE TABLE IF NOT EXISTS C (key TEXT PRIMARY KEY, val TEXT);"
sqlite3 config.db "INSERT OR REPLACE INTO C (key, val) VALUES ('http.addr', 'localhost:3001');"
sqlite3 config.db "INSERT OR REPLACE INTO C (key, val) VALUES ('approot.scheme', 'https');"
sqlite3 config.db "INSERT OR REPLACE INTO C (key, val) VALUES ('approot.host', 'localhost:3001');"
sqlite3 config.db "INSERT OR IGNORE INTO C (key, val) VALUES ('atom.id', '${ATOM_ID}');"

./atlas \
  -charts                "${HOME}/charts/" \
  -chartsroot            "" \
  -etherpadApiUrl        "http://localhost:9001/api" \
  -etherpadApiSecretPath "${HOME}/src/eplite/APIKEY.txt" \
  -alsologtostderr=true 2>&1

