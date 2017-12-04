IMG = gcr.io/cloud-ml-dev/test_launcher
TAG := $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)-$(shell git diff | sha256sum | cut -c -6)
DIR := ${CURDIR}

# 1.4 isn't new enough for the tf-benchmarks code
# so we pin to a particular nightly build image.
# CPU_BASE = tensorflow/tensorflow:nightly
CPU_BASE = tensorflow/tensorflow:@sha256:5edc0446cc989ad75bc30631f89f20694fe5bf5226f665d47e5c7f35a3b18484
# GPU_BASE = tensorflow/tensorflow:nightly-gpu
GPU_BASE = tensorflow/tensorflow@sha256:bfadad8f2c80424d8d6059d3b8cd6947bf23111dc786fc33db72b56b632a1f28

REGISTRY := gcr.io/cloud-ml-dev
BENCH_MARKS_IMAGE := $(REGISTRY)/tf-benchmarks

# Build the cpu image
build-cpu:
	jinja2 Dockerfile.template --format=yaml -D base_image=$(CPU_BASE) > Dockerfile.cpu
	docker build -t $(BENCH_MARKS_IMAGE)-cpu:$(TAG) -f Dockerfile.cpu ./
	gcloud docker -- push $(BENCH_MARKS_IMAGE)-cpu:$(TAG)

build-gpu:
	jinja2 Dockerfile.template --format=yaml -D base_image=$(GPU_BASE) > Dockerfile.gpu
	docker build -t $(BENCH_MARKS_IMAGE)-gpu:$(TAG) -f Dockerfile.gpu ./
	gcloud docker -- push $(BENCH_MARKS_IMAGE)-gpu:$(TAG)