#/bin/bash

set -x

pushd test
trap popd EXIT

echo "asdasd"

echo "!23123"

terraform plan -input=true -lock=false
