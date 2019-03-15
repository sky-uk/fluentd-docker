THIS_FILE := $(lastword $(MAKEFILE_LIST))
KIND_CLUSTER := "es-e2e"

define kind_kubeconfig
	$(shell kind get kubeconfig-path --name $(KIND_CLUSTER))
endef

# Docker build
git_rev := $(shell git rev-parse --short HEAD)
git_tag := $(shell git tag --points-at=$(git_rev))
image_prefix := skycirrus/fluentd-docker
image_latest := $(image_prefix):latest
image_file := fluentd-docker.tar.gz
kind_node_prefix := kind-$(KIND_CLUSTER)-

all: docker kind kind-docker-load-image kind-deploy-resources e2e kind-cleanup

.PHONY: docker
docker :
	@echo "== build docker images"
	docker build -t $(image_latest) .

.PHONY: kind-cleanup
kind-cleanup:
	@echo "== delete local KinD $(KIND_CLUSTER) cluster"
	kind delete cluster --name=$(KIND_CLUSTER)

.PHONY: kind
kind: kind-cleanup
	@echo "== create local KinD $(KIND_CLUSTER) cluster"
	# Create a new KinD cluster using a custom configuration
	kind create cluster --name=$(KIND_CLUSTER) --config e2e/kind.conf

.PHONY: kind-docker-load-image
kind-docker-load-image:
	@echo "== load image into KinD $(KIND_CLUSTER) cluster"
	@# Export generated Docker image to an archive.
	docker save $(image_latest) -o $(image_file)
	@# Copy saved archive into kind's Docker containers and
	@# Import image into their Docker daemon to make it accessible to Kubernetes.
	@for node in "control-plane" "worker" ; do \
		echo "- loading image '$(image_latest)' to $$node" ; \
		docker cp $(image_file) $(kind_node_prefix)$$node:/$(image_file) ; \
		docker exec $(kind_node_prefix)$$node docker load -i /$(image_file) ; \
	done
	@# Cleanup archive
	rm $(image_file)

.PHONY: kind-deploy-resources
kind-deploy-resources:
	@echo "== deploy k8s resources into KinD $(KIND_CLUSTER) cluster"
	@echo "-- elasticsearch"
	kubectl --kubeconfig $(call kind_kubeconfig) apply -f e2e/resources/es
	@echo "-- fluentd"
	kubectl --kubeconfig $(call kind_kubeconfig) apply -f e2e/resources/fluentd
	@echo "-- logging-pod"
	kubectl --kubeconfig $(call kind_kubeconfig) apply -f e2e/resources/logging-pod.yml

.PHONY: e2e
e2e: kind-deploy-resources
	@echo "== run end to end tests"

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
