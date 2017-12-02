IMG = gcr.io/cloud-ml-dev/test_launcher
TAG := $(shell date +v%Y%m%d)-$(shell git describe --tags --always --dirty)-$(shell git diff | sha256sum | cut -c -6)
DIR := ${CURDIR}

CPU_BASE = tensorflow/tensorflow:nightly

BENCH_MARKS_IMAGE := gcr.io/cloud-ml-dev/tf-benchmarks

# Build the cpu image
build-cpu:
	jinja2 Dockerfile.template --format=yaml -D base_image=$(CPU_BASE) > Dockerfile.cpu
	docker build -t $(BENCH_MARKS_IMAGE):$(TAG) -f Dockerfile.cpu ./
	gcloud docker -- push $(BENCH_MARKS_IMAGE):$(TAG)

build:
	docker build -t $(IMG):$(TAG) -f Dockerfile.example ./
	gcloud docker -- push $(IMG):$(TAG)
	jinja2 test_job.yaml.template --format=yaml -D image=$(IMG):$(TAG) > test_job.yaml

deploy: build
	# We ignore kubectl delete failures as it likely
	# indicates the job doesn't exist.
	-kubectl delete -f test_job.yaml
	kubectl create -f test_job.yaml