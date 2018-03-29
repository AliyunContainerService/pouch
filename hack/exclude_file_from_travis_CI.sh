#!/bin/bash

set -e

CHANGED_FILES=`git diff --name-only ${TRAVIS_COMMIT}`
SKIP_CI=True
MD=".md"
YML=".yml"

for CHANGED_FILE in $CHANGED_FILES; do
  if ! [[ $CHANGED_FILE =~ $MD ]] || ! [[ $CHANGED_FILE =~ $YML ]]; then
    SKIP_CI=False
    break
  fi
done

if [[ $SKIP_CI == True ]]; then
  echo "Only .md files found, exiting."
  travis cancel ${TRAVIS_BUILD_ID}
  exit 0
else
  echo "Non-.md files found, continuing with build."
fi
