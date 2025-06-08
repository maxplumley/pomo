#!/bin/bash

# Check if both arguments are provided
if [ $# -ne 1 ]; then
    echo "Error: Invalid number of arguments"
    echo "Usage: $0 <version>"
    exit 1
fi

# Store the arguments
VERSION="$1"

echo "Bumping version to: $VERSION"

echo "should_release=true" >> $GITHUB_OUTPUT
echo "new_version=${VERSION}" >> $GITHUB_OUTPUT

exit 0
