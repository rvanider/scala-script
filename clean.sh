#!/bin/bash

find . -name '*.g.*' -exec rm {} \;
find . -name '*.jar' -exec rm {} \;
rm -rf ./src/one/com
rm -rf ./src/two/com
