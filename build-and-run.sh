#!/usr/bin/env bash

docker build -t dialog/push-server .
#docker run --rm -it --log-driver=gelf --log-opt gelf-address=udp://logstash.transmit.im:12201  -p 8010:8010 -p 8011:8011 -v `pwd`:/config dialog/push-server -c /config/example.yaml
docker run --rm -it -p 8010:8010 -p 8011:8011 -v `pwd`:/config dialog/push-server -c /config/example.yaml -g logstash.transmit.im:12201
