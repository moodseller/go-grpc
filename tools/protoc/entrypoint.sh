#!/bin/sh
#
# This script compiles given proto files to given output dir.

set -e

inputPath='./api'
outputPath='./api'

compile() {
    for d in ${inputPath}/*/; do
        dir_name="$(basename "$d")"
        filename=${dir_name}.proto
        protoc --proto_path ./api/${dir_name} --go_out=plugins=grpc:./${outputPath}/${dir_name} ${filename}
    done
}

compile
