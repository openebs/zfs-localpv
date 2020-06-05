#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that directory
cd "$DIR"

# Get the git commit
if [ -f $PWD/GITCOMMIT ];
then
    GIT_COMMIT="$(cat $PWD/GITCOMMIT)"
else
    GIT_COMMIT="$(git rev-parse HEAD)"
fi

# Set BUILDMETA based on travis tag
if [[ -n "$TRAVIS_TAG" ]] && [[ $TRAVIS_TAG != *"RC"* ]]; then
    echo "released" > BUILDMETA
fi

CURRENT_BRANCH=""
if [ -z ${TRAVIS_BRANCH} ];
then
  CURRENT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2)
else
  CURRENT_BRANCH=${TRAVIS_BRANCH}
fi

# Get the version details
if [ -n "$TRAVIS_TAG" ]; then
	VERSION="$TRAVIS_TAG"
else
	BUILDDATE=`date +%m-%d-%Y`
	SHORT_COMMIT="$(git rev-parse --short HEAD)"
	VERSION="$CURRENT_BRANCH-$SHORT_COMMIT:$BUILDDATE"
fi

echo -e "\nbuilding the ZFS Driver version :- $VERSION\n"

VERSION_META="$(cat $PWD/BUILDMETA)"

# Determine the arch/os combos we're building for
UNAME=$(uname)
ARCH=$(uname -m)
if [ "$UNAME" != "Linux" -a "$UNAME" != "Darwin" ] ; then
    echo "Sorry, this OS is not supported yet."
    exit 1
fi

if [ "$UNAME" = "Darwin" ] ; then
  XC_OS="darwin"
elif [ "$UNAME" = "Linux" ] ; then
  XC_OS="linux"
fi

if [ "${ARCH}" = "i686" ] ; then
    XC_ARCH='386'
elif [ "${ARCH}" = "x86_64" ] ; then
    XC_ARCH='amd64'
else
    echo "Unusable architecture: ${ARCH}"
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
rm -rf bin/${PNAME}/*
mkdir -p bin/${PNAME}/

# If its dev mode, only build for ourself
if [[ "${DEV}" ]]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

# Build!
echo "==> Building ${CTLNAME} using $(go version)... "

GOOS="${XC_OS}"
GOARCH="${XC_ARCH}"
output_name="bin/${PNAME}/"$GOOS"_"$GOARCH"/"$CTLNAME

if [ $GOOS = "windows" ]; then
    output_name+='.exe'
fi
env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags \
    "-X github.com/openebs/zfs-localpv/pkg/version.GitCommit=${GIT_COMMIT} \
    -X main.CtlName='${CTLNAME}' \
    -X github.com/openebs/zfs-localpv/pkg/version.Version=${VERSION} \
    -X github.com/openebs/zfs-localpv/pkg/version.VersionMeta=${VERSION_META}"\
    -o $output_name\
    ./cmd

echo ""

# Move all the compiled things to the $GOPATH/bin
GOPATH=${GOPATH:-$(go env GOPATH)}
case $(uname) in
    CYGWIN*)
        GOPATH="$(cygpath $GOPATH)"
        ;;
esac
OLDIFS=$IFS
IFS=: MAIN_GOPATH=($GOPATH)
IFS=$OLDIFS

# Create the gopath bin if not already available
mkdir -p ${MAIN_GOPATH}/bin/

# Copy our OS/Arch to the bin/ directory
DEV_PLATFORM="./bin/${PNAME}/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
    cp ${F} bin/${PNAME}/
    cp ${F} ${MAIN_GOPATH}/bin/
done

if [[ "x${DEV}" == "x" ]]; then
    # Zip and copy to the dist dir
    echo "==> Packaging..."
    for PLATFORM in $(find ./bin/${PNAME} -mindepth 1 -maxdepth 1 -type d); do
        OSARCH=$(basename ${PLATFORM})
        echo "--> ${OSARCH}"

        pushd "$PLATFORM" >/dev/null 2>&1
        zip ../${PNAME}-${OSARCH}.zip ./*
        popd >/dev/null 2>&1
    done
fi

# Done!
echo
echo "==> Results:"
ls -hl bin/${PNAME}/
