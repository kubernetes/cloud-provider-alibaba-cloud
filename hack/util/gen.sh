#!/bin/bash

#controller-gen  crd paths=./pkg/apis/... output:crd:dir=deploy/crds output:stdout

controller-gen object paths=./pkg/apis/alibabacloud/v1/...