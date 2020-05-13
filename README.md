# point-to-point network driver plugin for docker

For efficient container-to-container networking a simple non-bridged veth pair
is sufficient. This docker network driver implements a point-to-point network
using a veth pair.

## Usage

To use with docker
```
docker network create --driver="choppsv1/docker-network-p2p" testnet
docker run --rm --network=testnet alpine ip addr
```

To use in docker compose:

```
version: "2.4"

services:
  h1:
    image: alpine
    command: sh -c 'ip addr; tail -f /dev/null'
    networks: [ "p2pnet" ]
  h2:
    image: alpine
    command: sh -c 'ip addr; tail -f /dev/null'
    networks: [ "p2pnet" ]

networks:
  p2pnet:
    driver: "choppsv1/docker-network-p2p"
```

# License
The software contained herein is licensed under Apache License (Version 2.0)
