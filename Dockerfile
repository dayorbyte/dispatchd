#
# Everything about this is kind of gross, but it does get a server running
#

FROM centos:latest
# OS setup
RUN yum install -y make golang git
RUN mkdir -p /app/dispatchd
RUN yum install -y python-setuptools.noarch
RUN easy_install mako
RUN yum install -y gcc-c++ glibc-headers

# protobuf
RUN cd /tmp && curl -L -o protobuf-2.6.1.tar.gz  https://github.com/google/protobuf/releases/download/v2.6.1/protobuf-2.6.1.tar.gz
RUN cd /tmp && tar -xzf protobuf-2.6.1.tar.gz
RUN cd /tmp/protobuf-2.6.1/ && ./configure && make install

# Build dispatchd
RUN mkdir -p /app/dispatchd/src/github.com/jeffjenkins/dispatchd/
COPY . /app/dispatchd/src/github.com/jeffjenkins/dispatchd/
ENV GOPATH /app/dispatchd
RUN cd /app/dispatchd/src/github.com/jeffjenkins/dispatchd/ && PATH=$PATH:$GOPATH/bin make install

# Runtime configuration
# TODO: when running for real this should have a volume for
#       the database directories
RUN mkdir -p /data/dispatchd/
ENV STATIC_PATH /app/dispatchd/src/github.com/jeffjenkins/dispatchd/static
ENTRYPOINT ["/app/dispatchd/bin/server"]
