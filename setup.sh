#!/bin/sh

set -xe

git clean -dfx

go build -o ./atlas akamai/atlas

URL="http://localhost:3001"

(sleep 1; xdg-open "$URL" || open "$URL") &

./atlas \
  -http                  "localhost:3001" \
  -charts                "${HOME}/charts/" \
  -chartsroot            "" \
  -etherpadApiUrl        "http://localhost:9001/api" \
  -etherpadApiSecretPath "${HOME}/src/eplite/APIKEY.txt"

