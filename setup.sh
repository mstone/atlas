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

(sleep 1; xdg-open 'http://localhost:3001' || open 'http://localhost:3001') &

./atlas-forms \
  -http       'localhost:3001' \
  -forms      'data/' \
  -formsroot  'forms' \
  -charts     "${HOME}/p4/docs/security/arch/" \
  -chartsroot ''

