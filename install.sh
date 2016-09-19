#!/bin/bash

curl=$(which curl)
wget=$(which wget)
dest=/usr/local/bin
url=https://github.com/rvanider/scala-script/blob/master/scala-script?raw=true

if [ -z "$curl" ] && [ -z "$wget" ]; then
  echo "unable to install, need curl or wget"
  exit 1
fi

mkdir -p $dest
if [ $? -ne 0 ]; then
  echo "unable to create $dest, permissions?"
  exit 1
fi

if [ -n "$curl" ]; then
  curl -s --fail -L --output "$dest/scala-script" "$url"
else
  wget -q -O "$dest/scala-script" "$url"
fi
if [ $? -ne 0 ]; then
  echo "unable to download $url"
  exit 1
fi

chmod +x "$dest/scala-script"

echo "scala-script installed to $dest"
echo "use #!/usr/bin/env scala-script as script header"
echo ""
