#!/bin/bash

mkdir /tmp/{working,finished}

echo Building Config store...
cd keyvaluestore
go build -o ../bin/config-store

echo Building Tasks store...
cd ../Database
go build -o ../bin/tasks-store

echo Building Images store...
cd ../Storage
go build -o ../bin/images-store

echo Building Master...
cd ../Master
go build -o ../bin/master

echo Building Worker...
cd ../Worker
go build -o ../bin/worker

echo Building Frontend...
cd ../Frontend
go build -o ../bin/frontend

echo Done.
