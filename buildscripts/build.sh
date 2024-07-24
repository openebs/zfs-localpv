#!/usr/bin/env bash

set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that directory
cd "$DIR"

# Get the git commit
if [ -f "$GOPATH"/src/github.com/openebs/zfs-localpv/GITCOMMIT ];
then
    GIT_COMMIT=$(cat "$GOPATH"/src/github.com/openebs/zfs-localpv/GITCOMMIT)
else
    GIT_COMMIT="$(git rev-parse HEAD)"
fi

# Set BUILDMETA based on travis tag
if [[ -n "$RELEASE_TAG" ]] && [[ "$RELEASE_TAG" != *"RC"* ]]; then
    echo "released" > BUILDMETA
fi

CURRENT_BRANCH=""
if [ -z "${BRANCH}" ];
then
  CURRENT_BRANCH=$(git branch | grep '\*' | cut -d ' ' -f2)
else
  CURRENT_BRANCH=${BRANCH}
fi

# Get the version details
if [ -n "$RELEASE_TAG" ]; then
	VERSION="$RELEASE_TAG"
else
	BUILDDATE="$(date +%m-%d-%Y)"
	SHORT_COMMIT="$(git rev-parse --short HEAD)"
	VERSION="$CURRENT_BRANCH-$SHORT_COMMIT:$BUILDDATE"
fi

echo -e "\nbuilding the ZFS Driver version :- $VERSION\n"

VERSION_META=$(cat "$PWD/BUILDMETA")

# Determine the arch/os combos we're building for
UNAME=$(uname)
if [ "$UNAME" != "Linux" ] && [ "$UNAME" != "Darwin" ] ; then
    echo "Sorry, this OS is not supported yet."
    exit 1
fi

if [ -z "${PNAME}" ];
then
    echo "Project name not defined"
    exit 1
fi

if [ -z "${CTLNAME}" ];
then
    echo "CTLNAME not defined"
    exit 1
fi

# Delete the old dir
echo "==> Removing old directory..."
rm -rf bin/"$PNAME"/*
mkdir -p bin/"$PNAME"/

XC_OS=$(go env GOOS)
XC_ARCH=$(go env GOARCH)

# Build!
echo "==> Building ${CTLNAME} using $(go version)... "

GOOS="${XC_OS}"
GOARCH="${XC_ARCH}"
output_name=bin/"$PNAME"/"$GOOS"_"$GOARCH"/$CTLNAME

if [ "$GOOS" = "windows" ]; then
    output_name+='.exe'
fi
env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build -mod=mod -ldflags \
    "-X github.com/openebs/zfs-localpv/pkg/version.GitCommit=${GIT_COMMIT} \
    -X main.CtlName='${CTLNAME}' \
    -X github.com/openebs/zfs-localpv/pkg/version.Version=${VERSION} \
    -X github.com/openebs/zfs-localpv/pkg/version.VersionMeta=${VERSION_META}"\
    -o "$output_name"\
    ./cmd

echo ""

# Move all the compiled things to the $GOPATH/bin
GOPATH=${GOPATH:-$(go env GOPATH)}
case $(uname) in
    CYGWIN*)
        GOPATH=$(cygpath "$GOPATH")
        ;;
esac
OLDIFS=$IFS
IFS=: MAIN_GOPATH="$GOPATH"
IFS=$OLDIFS

# Create the gopath bin if not already available
mkdir -p "${MAIN_GOPATH}"/bin/

# Copy our OS/Arch to the bin/ directory
DEV_PLATFORM=./bin/"${PNAME}"/$(go env GOOS)_$(go env GOARCH)
find "${DEV_PLATFORM}" -mindepth 1 -maxdepth 1 -type f -print0 | while IFS= read -r -d '' file
do
    cp "$file" bin/"${PNAME}"/
    cp "$file" "${MAIN_GOPATH}"/bin/
done

# Done!
echo
echo "==> Results:"
ls -hl bin/"${PNAME}"/
