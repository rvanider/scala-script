#!/bin/bash

./clean.sh
./build.sh
ret=$?
if [ "$ret" != "0" ]; then
  exit $ret
fi

echo testing
./scala-script test/test.scala
ret=$?
if [ "$ret" == "0" ]; then
  ./clean.sh
fi
exit $ret
