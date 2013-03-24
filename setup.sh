#!/bin/sh

set -xe

git clean -dfx

go build -o ./atlas-forms akamai/atlas/forms

#URL="http://localhost:3001/hades/design/system.svg/editor"
URL="http://localhost:3001"

(sleep 1; xdg-open "$URL" || open "$URL") &

./atlas-forms \
  -http                  "localhost:3001" \
  -charts                "${HOME}/p4/docs/security/arch/" \
  -chartsroot            "" \
  -etherpadApiUrl        "http://localhost:9001/api" \
  -etherpadApiSecretPath "${HOME}/src/eplite/APIKEY.txt"

