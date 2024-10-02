#! /bin/bash

# Purpose: To clone the repo

# 1. Clone the repo
export GITHUB_REPO_URL=$GITHUB_REPO_URL
git clone $GITHUB_REPO_URL /app/output

# 3. execute the build-server script
/app/bin/build-server
