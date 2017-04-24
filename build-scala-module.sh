#!/usr/bin/env bash

mkdir -p src/main/protobuf
cp -f src/protobuf/*.proto src/main/protobuf

grep -rl 'package server' src/main/protobuf | xargs sed -i.backup '/package server/a\
\
import "scalapb/scalapb.proto";\
option (scalapb.options) = {\
package_name: "im.dlg.push.service";\
};\
'
rm src/main/protobuf/*.proto.backup

sbt compile publish
