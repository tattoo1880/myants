#!/bin/zsh

git add .
git commit -m "$1"
git push origin main

git tag -a v0.1.0 -m "v0.1.0"

git push origin v0.1.0

git tag


git ls-remote --tags origin