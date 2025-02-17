#!/bin/bash
build_docker() {
    docker build "${ROOT}" \
        --file ./docker/dockerfile \
        --tag="${DOCKER_IMG_TAG%:*}:${VERSION}" \
        --build-arg="GOOS=${GOOS}" \
        --build-arg="GOARCH=${GOARCH}"
}

run_docker_multi_arch_build() {
    docker run --rm -it \
        -e="VERSION=${VERSION}" \
        --workdir="/work" \
        --entrypoint="/bin/sh" \
        -v="$PWD:/data:ro" \
        -v="$PWD/bin:/build" \
            "${DOCKER_IMG_MULTI_ARCH_TAG}" \
            -c /data/docker/multi_arch_build.sh
}

run_docker_test() {
    docker run --rm -it \
        --workdir="/work" \
        --tmpfs="/work/internal/interpreter_new/testdata/imports/bin/" \
        -v="$PWD:/work:ro" \
            "${DOCKER_IMG_MULTI_ARCH_TAG}" \
            go test /work/internal/interpreter_new/...
}