# Derived this from Makefile in https://github.com/vieux/docker-volume-sshfs
# which is MIT licensed.
DOCKER_REPO ?= choppsv1/docker-network-p2p
DOCKER_TAG ?= latest
IMAGE_NAME ?= ${DOCKER_REPO}:${DOCKER_TAG}

SOURCES := main.go $(wildcard */*.go)

all: clean create

clean:
	rm -rf ./plugin

# Build the image, then extract into a "rootfs" for docker plugin
plugin/rootfs/docker-network-p2p: $(SOURCES)
	docker build -q -t $(DOCKER_REPO):rootfs .
	docker create --name tmp $(DOCKER_REPO):rootfs
	mkdir -p ./plugin/rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp

create: plugin/rootfs/docker-network-p2p
	docker plugin rm -f $(IMAGE_NAME) || true
	docker plugin create $(IMAGE_NAME) ./plugin

enable:
	docker plugin enable $(IMAGE_NAME)

push:
	docker plugin push $(IMAGE_NAME)

test:
	hooks/test
