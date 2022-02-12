#!/usr/bin/env bash

export PROGRAM_NAME="elc"
export OWNER="MadridianFox"
export REPO="ensi-local-ctl"
export BIN_LOCATION="/opt/elc"
export LINK_LOCATION="/usr/local/bin"

version=$(curl -sI https://github.com/$OWNER/$REPO/releases/latest | grep -i "location:" | awk -F"/" '{ printf "%s", $NF }' | tr -d '\r')

if [ ! $version ]; then
  echo "Failed while attempting to install $REPO. Please manually install:"
  echo ""
  echo "1. Open your web browser and go to https://github.com/$OWNER/$REPO/releases"
  echo "2. Download the latest release for your platform. Call it '$PROGRAM_NAME-<version>'."
  echo "3. chmod +x ./$PROGRAM_NAME-<version>"
  echo "4. mv ./$PROGRAM_NAME-<version> $BIN_LOCATION"
  echo "5. ln -sf $BIN_LOCATION/$PROGRAM_NAME-<version> $LINK_LOCATION/$PROGRAM_NAME"

  exit 1
fi

targetFile="/tmp/elc_linux_amd64"
if [ -e "$targetFile" ]; then
  rm "$targetFile"
fi

if ! [ -e "$BIN_LOCATION" ]; then
  mkdir -p "$BIN_LOCATION"
fi

echo "Downloading package $url as $targetFile"
url="https://github.com/$OWNER/$REPO/releases/download/$version/elc_linux_amd64"
curl -sSL $url --output "$targetFile"

if [ "$?" = "0" ]; then
  echo "Download complete."
  chmod +x "$targetFile"
  mv "$targetFile" "$BIN_LOCATION/$PROGRAM_NAME-$version"
  if [ "$?" != "0" ]; then
      echo "Unable to move $targetFile into $BIN_LOCATION/$PROGRAM_NAME-$version"
      exit 1
  fi
  ln -s -f "$BIN_LOCATION/$PROGRAM_NAME-$version" "$LINK_LOCATION/$PROGRAM_NAME"
else
  echo "Unable to download $url into $targetFile"
  exit 1
fi

echo "Successfully installed"