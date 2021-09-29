#!/bin/bash

docker build . -t $(cat $1/meta/IMAGENAME) \
	--build-arg http_proxy="http://192.168.1.21:41091" \
	--build-arg https_proxy="http://192.168.1.21:41091" \
	-f $1/Dockerfile