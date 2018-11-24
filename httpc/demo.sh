#!/usr/bin/env bash

python httpc.py get "http://localhost:8080/httpfs.txt" -udp

python httpc.py post localhost:8080/bar.txt -d "I love food" -udp

python httpc.py post "http://localhost:8080/bar.txt" -d "food" -h "Content-Type:application/text" -h "Content-Length:4" -v -udp

python httpc.py post localhost:8080/hello.txt -d "I love food" -udp