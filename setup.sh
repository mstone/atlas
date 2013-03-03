#!/bin/sh

set -xe

git clean -dfx

if ! [ -f data/reviews.json ]; then
  cp data/questions.json.ex data/questions.json
fi

if ! [ -f data/profiles.json ]; then
  cp data/profiles.json.ex.catechism data/profiles.json
fi

if ! [ -f data/reviews.json ]; then
  cp data/reviews.json.ex data/reviews.json
fi

go build -o ./atlas-forms akamai/atlas/forms

./atlas-forms \
  -http       'localhost:3001' \
  -formsroot  'forms' \
  -charts     "${HOME}/p4/docs/security/arch/" \
  -chartsroot ''

