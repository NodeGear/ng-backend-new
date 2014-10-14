#!/bin/bash

export CONFIG_PATH=`pwd`/configuration.json

if [ "$1" == "test" ]
then
	go test nodegear/*
	exit 1
fi

if [ "$1" == "run" ]
then
	go run main.go
	exit 0
fi

if [ "$1" == "clean" ]
then
	rm -rf chroot/
	exit 0
fi

echo "Usage: arg1 = test/run/clean"
