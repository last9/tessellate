#! /bin/bash
# Get or Install protoc

# make bash more robust.
set -beEux -o pipefail

VERSION="${PROTO_VERSION:=3.5.1}"
PROTO_DIR="/tmp/proto/$PROTO_VERSION"

if [ ! -f "$PROTO_DIR/bin/protoc" ]; then
  mkdir -p ${PROTO_DIR}
  wget "https://github.com/google/protobuf/releases/download/v${PROTO_VERSION}/protoc-${PROTO_VERSION}-linux-x86_64.zip" -O /tmp/proto.zip
  unzip /tmp/proto.zip -d ${PROTO_DIR}
  rm /tmp/proto.zip
fi
