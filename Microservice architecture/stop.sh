#!/bin/bash

sudo pkill frontend
pkill worker
pkill master
pkill images-store
pkill tasks-store
pkill config-store

echo Done.
