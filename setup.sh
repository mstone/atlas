#!/bin/sh

set -xe

git clean -dfx

if ! [ -f data/records.json ]; then
  cp data/questions.json.ex data/questions.json
fi

if ! [ -f data/forms.json ]; then
  cp data/forms.json.ex.catechism data/forms.json
fi

if ! [ -f data/records.json ]; then
  cp data/records.json.ex data/records.json
fi

go build -o ./atlas-forms akamai/atlas/forms

#URL="http://localhost:3001/hades/design/system.svg/editor"
URL="http://localhost:3001"

(sleep 1; xdg-open "$URL" || open "$URL") &

./atlas-forms \
  -http       'localhost:3001' \
  -forms      'data/' \
  -formsroot  'forms' \
  -charts     "${HOME}/p4/docs/security/arch/" \
  -chartsroot ''

