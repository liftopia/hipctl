#!/bin/bash

if [ "$0" != "-bash" ]; then
  echo "you need to source me homie. (. $0)"
else
  echo "gettin you all set up!"
  export REDIS_URL=redis://127.0.0.1:6379/2
  PROG=hipctl source ~/go/src/github.com/codegangsta/cli/autocomplete/bash_autocomplete
fi
