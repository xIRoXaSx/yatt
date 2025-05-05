#!/bin/bash

build_docker() {
    docker build "${ROOT}" \
        --file ./docker/dockerfile \
        --tag="${DOCKER_IMG_TAG%:*}:${VERSION}" \
        --build-arg="GOOS=${GOOS}" \
        --build-arg="GOARCH=${GOARCH}"
}

run_docker_build_release() {
    docker run --rm -it \
        -e="VERSION=${VERSION}" \
        --workdir="/work" \
        --entrypoint="/bin/sh" \
        -v="$PWD:/data:ro" \
        -v="$PWD/bin:/build" \
            "${DOCKER_IMG_TAG_BUILD}" \
            -c /data/docker/build-release.sh
}

run_tests() {
    docker run --rm -it \
        --workdir="/work" \
        --tmpfs="/work/internal/interpreter/testdata/interpret/out" \
        -v="$PWD:/work:ro" \
            "${DOCKER_IMG_TAG_BUILD}" \
            go test \
                /work/internal/common/... \
                /work/internal/interpreter/...  \
                /work/internal/core/...
}