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

doit curl -i $APP/reviews/

doit curl -i -X POST -d '{}' $APP/reviews/

