.PHONY: all docker

all : docker
travis : docker

# Docker build
git_rev := $(shell git rev-parse --short HEAD)
git_tag := $(shell git tag --points-at=$(git_rev))
image_prefix := skycirrus/fluentd-docker

docker :
	@echo "== build docker images"
	docker build -t $(image_prefix):latest .

release : docker
	@echo "== release docker images"
ifeq ($(strip $(git_tag)),)
	@echo "no tag on $(git_rev), skipping release"
else
	@echo "releasing $(image):$(git_tag)"
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	docker tag $(image_prefix):latest $(image_prefix):$(git_tag)
	docker push $(image_prefix):$(git_tag)
	docker push $(image_prefix):latest
endif
