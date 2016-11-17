#!/bin/bash

# Generate changelog
# uses https://github.com/skywinder/github-changelog-generator

github_changelog_generator -b CHANGELOG.md \
    --bugs-label="### Bugs" \
    --enhancement-label="### Enhancements" \
    --issues-label="### Closed Issues" \
    --pr-label="### Closed Pull Requests" 
