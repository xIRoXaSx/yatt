#!/bin/bash
apt update && apt install -y zip

cd /data
os=("windows" "linux" "darwin" "freebsd" "openbsd" "windows")
arch=("386" "amd64" "arm" "arm64")
binName="fastplate"

for o in "${os[@]}" ; do
  for a in "${arch[@]}" ; do
    # Skip certain os - architecture pairs.
    [[ "${o}" == "freebsd" || "${o}" == "openbsd" && "${a}" == "arm64" ]] && continue
    [[ "${o}" == "windows" && "${a}" == "arm"* ]] && continue

    # Add extension for windows.
    ext=""
    [[ "${o}" == "windows" ]] && ext=".exe"

    echo "Building ${o}/${a}..."
    out="${binName}${ext}"
    CGO_ENABLED=0 GOOS=${o} GOARCH=${a} go build -buildvcs=false -ldflags "-w -extldflags '-static'" -o "/work/${out}"
    echo "Zipping..."
    zip -j "/build/${binName}-${VERSION}_${o}_${a}.zip" "/work/${out}"
    echo "---"
  done
done
