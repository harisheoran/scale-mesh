#! /bin/bash

# Purpose:

# 1. Clone the repo
read GITHUB_REPO_URL
git clone $GITHUB_REPO_URL webapp

# 2. Build the app
cd webapp && npm install && npm run build

# 3. execute script
go run build.go
