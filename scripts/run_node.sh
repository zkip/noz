#!/bin/bash

cd $(dirname $0)
cd "../packages/$1"

if [ $2 = build ]; then
	../../scripts/build-image-node.sh $1
fi

if [ $2 = push ]; then
	../../scripts/push-image.sh
fi

if [ $2 = deploy ]; then
	../../scripts/apply-to-cluster.sh
fi

if [ $2 = upgrade ]; then
	echo "===========build============"
	../../scripts/build-image.sh
	echo "===========push============"
	../../scripts/push-image.sh
	echo "===========deploy============"
	../../scripts/apply-to-cluster.sh
fi
