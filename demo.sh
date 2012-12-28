#!/bin/bash

doit() { 
   echo "$@"
   "$@"
   echo
   echo
   echo
}

APP=http://localhost:3001
doit curl -i -X POST -d '{}' $APP/reviews/foo

doit curl -i $APP/

doit curl -i $APP/profiles/

doit curl -i -X POST -d 'profile=pace-1.0.0' $APP/reviews/

doit curl -i $APP/reviews/

# doit curl -i -X POST -d '{}' $APP/reviews/
doit curl -i -X POST -d 'profile=pace-1.0.0&review=acme-1.0.0' $APP/reviews/

doit curl -i $APP/reviews/
