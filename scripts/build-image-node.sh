#!/bin/bash

docker build . -t $(cat meta/IMAGENAME) \
	--build-arg http_proxy="http://192.168.1.21:41091" \
	--build-arg https_proxy="http://192.168.1.21:41091"