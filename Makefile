REPO = 889883130442.dkr.ecr.us-west-2.amazonaws.com/k8-cross-cluster-controller
VERSION := $(shell git rev-parse --short HEAD)

build:
	docker build -t k8-cross-cluster-controller:$(VERSION) .

release:
	docker tag k8-cross-cluster-controller:$(VERSION) $(REPO):$(VERSION) && \
		docker push $(REPO):$(VERSION)
