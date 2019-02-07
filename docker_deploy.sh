#! /bin/bash
NAME=$1
if [ -n $CIRCLE_TAG ]; then
    CIRCLE_TAG=$(git describe HEAD --tags)
fi
VERSION=$(echo $CIRCLE_TAG | sed -E 's/^v//')

get_component () {
    echo $VERSION | cut -f ${1} -d '.'
}

MAJOR=$(get_component 1)
MINOR=$(get_component 2)
PATCH=$(get_component 3 | cut -f 1 -d '-')

TAGS=()
if [[ "$VERSION" = *-* ]]; then 
    PRERELEASE_TAG=$(get_component 3 | cut -f 2 -d '-')
    TAGS+=("$PRERELEASE_TAG")
    TAGS+=("$VERSION")
else
    TAGS+=("latest")
    TAGS+=("$MAJOR")
    TAGS+=("$MAJOR.$MINOR")
    TAGS+=("$MAJOR.$MINOR.$PATCH")
fi

echo "## Identified tags for this version"
for t in ${TAGS[@]}; do echo $NAME:$t; done

echo ""
echo "## Building docker image"

BUILD_TAGS=()
for t in ${TAGS[@]}; do BUILD_TAGS+=("-t $NAME:$t"); done

docker build ${BUILD_TAGS[@]} .

echo ""
echo "## Pushing docker image"

for t in ${TAGS[@]}; do docker push $NAME:$t; done
