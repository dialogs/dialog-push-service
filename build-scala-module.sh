#!/usr/bin/env bash

mkdir -p src/main/protobuf
cp -f src/protobuf/*.proto src/main/protobuf

grep -rl 'package main' src/main/protobuf | xargs sed -i.backup '/package main/a\
\
import "scalapb/scalapb.proto";\
option (scalapb.options) = {\
package_name: "im.dlg.push.service";\
};\
'
rm src/main/protobuf/*.proto.backup

sbt clean compile publish-local
