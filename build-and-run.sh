#!/usr/bin/env bash

docker build -t dialog/push-server .
docker run --rm -it -p 8010:8010 -p 8011:8011 -v `pwd`:/config dialog/push-server -c /config/example.yaml
