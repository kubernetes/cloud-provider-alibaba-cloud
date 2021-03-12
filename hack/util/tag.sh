#!/usr/bin/env bash

echo "$(git describe --tags|awk -F "-" '{print $1}')-$(git rev-parse --short HEAD)"
