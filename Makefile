# Derived this from Makefile in https://github.com/vieux/docker-volume-sshfs
# which is MIT licensed.
PLUGIN_NAME = choppsv1/docker-network-p2p
PLUGIN_TAG ?= 1.0

SOURCES=main.go $(wildcard */*.go)

all: clean create

clean:
	rm -rf ./plugin

# Build the image, then extract into a "rootfs" for docker plugin
plugin/rootfs/docker-network-p2p: $(SOURCES)
	docker build -q -t ${PLUGIN_NAME}:rootfs .
	docker create --name tmp ${PLUGIN_NAME}:rootfs
	mkdir -p ./plugin/rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp

create: plugin/rootfs/docker-network-p2p
	docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./plugin

enable:
	docker plugin enable ${PLUGIN_NAME}:${PLUGIN_TAG}

push:  clean create
	docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}
