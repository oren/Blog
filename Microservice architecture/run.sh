#!/bin/bash

cd bin

echo Run Config store...
./config-store &
sleep 2

echo Run Tasks store...
./tasks-store 127.0.0.1:3001 127.0.0.1:3000 &

echo Run Image store...
./images-store 127.0.0.1:3002 127.0.0.1:3000 &

echo Run Master...
./master 127.0.0.1:3003 127.0.0.1:3000 &
sleep 3

echo Run Worker...
./worker 127.0.0.1:3000 3 &

echo Frontend...
./frontend 127.0.0.1:3000 &

echo Done.
