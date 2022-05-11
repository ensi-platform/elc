#!/usr/bin/env bash

tag=$(git describe --tags --abbrev=0 | sed 's/^v//')
commits=$(git rev-list v$tag..HEAD --count)

if [ "$commits" != "0" ]; then
  patch=$(echo $tag | sed -E 's/^[0-9]+.[0-9]+.([0-9]+)/\1/')
  prefix=$(echo $tag | sed -E 's/^([0-9]+.[0-9]+).[0-9]+/\1/')
  betaPath=$(($patch + 1))
  version="$prefix.$betaPath-beta.$commits"
else
  version=$tag
fi

echo "$version"