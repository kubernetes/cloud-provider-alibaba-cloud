#!/bin/bash

controller-gen  crd paths=./pkg/apis/... output:crd:dir=deploy/crds output:stdout
