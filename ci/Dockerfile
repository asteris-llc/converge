FROM golang:1.7

# update everything
RUN apt-get update

# install graphviz
RUN apt-get install -y graphviz

# install s3cmd
RUN apt-get install -y s3cmd

# install protoc
RUN apt-get install -y unzip && \
    wget https://github.com/google/protobuf/releases/download/v3.0.0/protoc-3.0.0-linux-x86_64.zip && \
    unzip protoc-3.0.0-linux-x86_64.zip && \
    mv bin/* /usr/bin/ && \
    mv include/google /usr/include/google && \
    go get -v github.com/golang/protobuf/protoc-gen-go github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger && \
    apt-get purge -y unzip

# install gox
RUN go get -v github.com/mitchellh/gox

WORKDIR /go
CMD /bin/bash
