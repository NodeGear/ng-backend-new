#!/bin/bash

set -o errexit

go get gopkg.in/mgo.v2
go get github.com/garyburd/redigo/redis
