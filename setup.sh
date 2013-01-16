#!/bin/sh

git clean -dfx

(cd src && go get)

if ! [ -f data/profiles.json ]; then
  cp data/profiles.json.ex.catechism data/profiles.json
fi

if ! [ -f data/reviews.json ]; then
  cp data/reviews.json.ex data/reviews.json
fi

go run src/main.go
