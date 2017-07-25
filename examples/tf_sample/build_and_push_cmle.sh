#!/bin/bash
#
# Script to build the tf_sample and push it to GCS.

set -e

SRC_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR=${SRC_DIR}/../../

. ${ROOT_DIR}/config.sh

IMAGE=gcr.io/cloud-ml-dev/tf_smoke_cmle:latest
SRC_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker build -t $IMAGE -f ${SRC_DIR}/Dockerfile.cmle.gpu ${SRC_DIR}
gcloud docker -- push $IMAGE
echo pushed $IMAGE
