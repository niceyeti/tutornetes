#!/bin/bash

cd code
docker build -f Dockerfile . -t simple-go-app
