FROM golang:1.14.2 as build
RUN go get github.com/docker/go-plugins-helpers/network
WORKDIR /go/src/github.com/choppsv1/docker-network-p2p
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-network-p2p .

FROM alpine
RUN apk update && \
        apk add iproute2 && \
        mkdir -p /run/docker/plugins
COPY --from=build /docker-network-p2p /docker-network-p2p
CMD ["/docker-network-p2p"]
