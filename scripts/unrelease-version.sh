#!/bin/bash

echo "$1"
gh release delete $1 -y
git push --delete origin $1