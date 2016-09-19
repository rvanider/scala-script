#!/bin/bash

function build_lib
{
  folder=$1
  cd $folder
  rm -rf com
  scalac *.scala
  mkdir -p ../../test/lib
  rm -f ../../test/lib/$folder.jar
  zip -r ../../test/lib/$folder.jar META-INF com
  rm -rf com
  cd ..
}

cd src
build_lib one
build_lib two
