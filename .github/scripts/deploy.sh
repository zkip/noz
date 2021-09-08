#!/bin/bash

mkdir -p docker-config
mv docker-config.json docker-config/config.json
docker --config docker-config pull $0

