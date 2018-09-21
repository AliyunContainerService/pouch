#!/bin/bash

set -o errexit
set -o nounset

#
# This script is used to regenerate api.pb.go.
#

# Get the absolute path of this file
DIR="$( cd "$( dirname "$0"  )" && pwd  )"/../..
API_ROOT="${DIR}/cri/apis/v1alpha2"

if [[ -z "$(which protoc)" || "$(protoc --version)" != "libprotoc 3."* ]]; then
  echo "Generating protobuf requires protoc 3.0.0-beta1 or newer. Please download and"
  echo "install the platform appropriate Protobuf package for your OS: "
  echo
  echo "  https://github.com/google/protobuf/releases"
  echo
  echo "WARNING: Protobuf changes are not being validated"
  exit 1
fi

protoc::install_gen_gogo(){
go get -u k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo
if ! which protoc-gen-gogo >/dev/null; then
  echo "GOPATH is not in PATH"
  exit 1
fi
}

protoc::install_gen_doc(){
go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc
if ! which protoc-gen-doc >/dev/null; then
  echo "GOPATH is not in PATH"
  exit 1
fi
}

protoc::generatedoc(){
    protoc::install_gen_doc
    protoc \
        --proto_path="${API_ROOT}" \
        --proto_path="${DIR}/vendor" \
        --doc_out="${DIR}/docs/api" \
        --doc_opt="${DIR}/hack/protoc/html.tmpl,CRI_API.html" "${API_ROOT}/api.proto"
}

protoc::generateproto(){
    protoc::install_gen_gogo
    protoc \
        --proto_path="${API_ROOT}" \
        --proto_path="${DIR}/vendor" \
        --gogo_out=plugins=grpc:"${API_ROOT}" "${API_ROOT}/api.proto"

    # Update boilerplate for the generated file.
    cat "${DIR}/hack/boilerplate/boilerplate.go.txt" "${API_ROOT}/api.pb.go" > "${API_ROOT}/tmpfile" \
        && mv "${API_ROOT}/tmpfile" "${API_ROOT}/api.pb.go"
    gofmt -l -s -w "${API_ROOT}/api.pb.go"
}

main(){
    local operation
    operation=$1

    if [[ "${operation}" == "gen_proto" ]]; then
        protoc::generateproto
    elif [[ "${operation}" == "gen_doc" ]]; then
        protoc::generatedoc
    else
        echo "invaild argument"
    fi
}
main "$@"
