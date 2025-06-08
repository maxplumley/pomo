#!/bin/bash

# Check if argument is provided
if [ $# -eq 0 ]; then
    echo "Error: No version number provided"
    echo "Usage: $0 <version>"
    exit 1
fi

# Store the version number
VERSION="$1"

echo "Bumping version to: $VERSION"
