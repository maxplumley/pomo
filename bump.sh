#!/bin/bash

# Check if both arguments are provided
if [ $# -ne 2 ]; then
    echo "Error: Invalid number of arguments"
    echo "Usage: $0 <version> <should_release>"
    exit 1
fi

# Store the arguments
VERSION="$1"
SHOULD_RELEASE="$2"

echo "Bumping version to: $VERSION"
echo "Release status: $SHOULD_RELEASE"

echo "should_release=true" >> $GITHUB_OUTPUT
echo "new_version=${VERSION}" >> $GITHUB_OUTPUT

exit 0
