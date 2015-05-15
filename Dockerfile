FROM golang:1.4

# cache-fill before COPY
# (docker/docker via git clone --depth=1 for speed)
RUN mkdir -p /go/src/github.com/docker \
	&& git clone --depth=1 https://github.com/docker/docker.git /go/src/github.com/docker/docker
ENV GOPATH $GOPATH:/go/src/github.com/docker/docker/vendor
# (also, naughty vbatts using things outside pkg/ from docker/docker)
RUN cd /go/src/github.com/docker/docker \
	&& bash hack/make/.go-autogen
RUN go get -d -v github.com/Sirupsen/logrus

WORKDIR /go/src/github.com/vbatts/docker-utils
COPY . /go/src/github.com/vbatts/docker-utils
RUN go get -d -v ./...
RUN go install -v ./...
