#!/bin/bash

doit() { 
   echo "$@"
   "$@"
   echo
   echo
   echo
}

APP=http://localhost:3001
doit curl -i -X POST -d '{}' $APP/records/foo

doit curl -i $APP/

doit curl -i $APP/forms/

doit curl -i -X POST -d 'form=pace-1.0.0' $APP/records/

doit curl -i $APP/records/

# doit curl -i -X POST -d '{}' $APP/records/
doit curl -i -X POST -d 'form=pace-1.0.0&record=acme-1.0.0' $APP/records/

doit curl -i $APP/records/
