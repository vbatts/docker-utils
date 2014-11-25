FROM centos

RUN yum install -y tar && yum clean all
RUN curl -o go.tar.gz https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz && tar xzf go.tar.gz
ENV GOROOT /go
ENV GOPATH /gopath
ENV PATH ${GOROOT}/bin:${GOPATH}/bin:${PATH}

RUN yum install -y git && yum clean all && \
	mkdir -p ${GOPATH}/src/github.com/vbatts

RUN git clone -b 1.0.1 git://github.com/vbatts/docker-utils.git ${GOPATH}/src/github.com/vbatts/docker-utils
RUN git clone -b v1.3.1 git://github.com/docker/docker.git ${GOPATH}/src/github.com/docker/docker
ENV GOPATH ${GOPATH}:${GOPATH}/src/github.com/docker/docker/vendor
RUN go install github.com/vbatts/docker-utils/cmd/dockertarsum

#CMD dockertarsum
