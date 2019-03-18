# Docker build
git_rev := $(shell git rev-parse --short HEAD)
git_tag := $(shell git tag --points-at=$(git_rev))
image_prefix := skycirrus/fluentd-docker
image_latest := $(image_prefix):latest

all: docker e2e

.PHONY: docker
docker :
	@echo "== build docker images"
	docker build -t $(image_latest) .

.PHONY: e2e-setup
e2e-setup:
	@echo "== setup"
	go get -u g

.PHONY: e2e
e2e: e2e-setup
	@echo "== run end to end tests"
	ginkgo -v e2e

release : docker
	@echo "== release docker images"
ifeq ($(strip $(git_tag)),)
	@echo "no tag on $(git_rev), skipping release"
else
	@echo "releasing $(image):$(git_tag)"
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	docker tag $(image_latest) $(image_prefix):$(git_tag)
	docker push $(image_prefix):$(git_tag)
	docker push $(image_latest)
endif
