#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Error: Invalid number of arguments"
    echo "Usage: $0 <version> <release_notes>"
    exit 1
fi

VERSION="$1"
RELEASE_NOTES="$2"

echo "Bumping version to: $VERSION"
echo "Release notes: $RELEASE_NOTES"

echo "should_release=true" >> $GITHUB_OUTPUT
echo "new_version=${VERSION}" >> $GITHUB_OUTPUT
echo "${RELEASE_NOTES}" >> /tmp/release_notes.md

exit 0
