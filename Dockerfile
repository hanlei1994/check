FROM golang AS build-env
COPY . /root

WORKDIR /root

RUN \
  export GO111MODULE=on &&\
  export GOPROXY=https://goproxy.io &&\
  CGO_ENABLED=0 go build -v 

FROM alpine

COPY --from=build-env /root/check /root/conf.toml / 
CMD ["/check"]

