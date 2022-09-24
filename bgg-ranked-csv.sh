#!/bin/bash

set -e

OUTPUT_FILE=$(date --iso-8601).csv
# We embed the source under /data in case we want to mount in
SOURCE_DIR=/data/bgg-ranking-historicals

# @see https://git-scm.com/docs/gitfaq#http-credentials-environment
git config --global credential.helper \
	'!f() { echo username=$GIT_CREDENTIAL_USERNAME; echo "password=$GIT_CREDENTIAL_PASSWORD"; };f'
git config --global user.name "$GIT_USER_NAME"
git config --global user.email "$GIT_USER_EMAIL"

mkdir -p /data

if [ -d "$SOURCE_DIR" ]; then
  # We mounted the repo in, so just clean and pull
  cd "$SOURCE_DIR"
  git stash -u
  git pull --no-rebase origin master
else
  # No repo yet, clone it
  git clone --depth=1 https://github.com/beefsack/bgg-ranking-historicals.git "$SOURCE_DIR"
fi
  
cd "$SOURCE_DIR"
bgg-ranked-csv > "$OUTPUT_FILE"
git add "$OUTPUT_FILE"
git commit -m "Added $OUTPUT_FILE"
git push origin master
