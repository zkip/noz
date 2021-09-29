#!/bin/bash

unset http_proxy
unset https_proxy

echo $(pwd)

kubectl set image deployments/$(cat meta/DEPLOYNAME) $(cat meta/CONTAINERNAME)=$(cat meta/IMAGENAME)
