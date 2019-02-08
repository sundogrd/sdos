#!/bin/bash

# The basename of our binary
BASE="sdos"

# Get the dependencies
go get -t -v -d $(go list ./...)


#
# We build on multiple platforms/archs
#
BUILD_PLATFORMS="linux windows darwin freebsd"
BUILD_ARCHS="amd64 386"

# For each platform
for OS in ${BUILD_PLATFORMS[@]}; do

    # For each arch
    for ARCH in ${BUILD_ARCHS[@]}; do

        # Setup a suffix for the binary
        SUFFIX="${OS}"

        # i386 is better than 386
        if [ "$ARCH" = "386" ]; then
            SUFFIX="${SUFFIX}-i386"
        else
            SUFFIX="${SUFFIX}-${ARCH}"
        fi

        # Windows binaries should end in .EXE
        if [ "$OS" = "windows" ]; then
            SUFFIX="${SUFFIX}.exe"
        fi

        echo "Building for ${OS} [${ARCH}] -> ${BASE}-${SUFFIX}"

        # Run the build
        export GOARCH=${ARCH}
        export GOOS=${OS}
        export CGO_ENABLED=0

        go build -ldflags "-X main.version=$(git describe --tags)" -o "${BASE}-${SUFFIX}" -tags=${OS} .

    done
done
