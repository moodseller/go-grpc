#!/bin/bash

grpcurl \
    --plaintext \
    -d '{}' \
    127.0.0.1:10000 user.User/GetUsers