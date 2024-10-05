#! /bin/bash
#############################
# Purpose: To clone the repo
#############################

export GITHUB_REPO_URL=$GITHUB_REPO_URL
export projectID=$projectID

# 1. Clone the repo
git clone $GITHUB_REPO_URL /app/output

# 3. execute the build-server script
/app/bin/build-server
