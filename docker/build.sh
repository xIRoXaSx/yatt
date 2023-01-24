#!/bin/bash
build_docker() {
  . ./.env
  sudo docker build ../ --file ./dockerfile \
     -t="${IMG_TAG}" \
     --build-arg="GOOS=${GOOS}" \
     --build-arg="GOARCH=${GOARCH}"
}

build_multi_arch_docker() {
  . ./.env
  sudo docker build . --file ./multi_arch_builder.dockerfile \
     -t="${IMG_MULTI_ARCH_TAG}" \
     --build-arg="GOOS=${GOOS}" \
     --build-arg="GOARCH=${GOARCH}"
}

run_docker_cli() {
  sudo docker run --rm -it fastplate:latest
}

run_docker_multi_arch_build() {
  . ./docker/.env
  sudo docker run --rm -it -e="VERSION=${VERSION}" -v="$PWD:/data:ro" -v="$PWD/bin:/build" -v="$PWD/docker/multi_arch_build.sh:/work/multi_arch_build.sh" "${IMG_MULTI_ARCH_TAG}"
}