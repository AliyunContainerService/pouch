#!/bin/bash

# Here is details of the swagger binary we use.
# $ swagger version
# version: 0.12.0
# commit: 8135eb6728e43b73489e80f94426e6d387809502

# Get the absolute path of this file
DIR="$( cd "$( dirname "$0"  )" && pwd  )"

swagger generate model -f $DIR/../apis/swagger.yml -t $DIR/../apis -m types
