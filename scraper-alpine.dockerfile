FROM alpine:3.6
MAINTAINER Rosco Pecoltran <https://github.com/roscopecoltran>

# build: docker build -t scraper:alpine -f scraper-alpine.dockerfile --no-cache .
# run: docker run --rm -ti -p 3000:3000 -v `pwd`:/app scraper:alpine

ARG GOPATH=${GOPATH:-"/go"}
ARG APK_INTERACTIVE=${APK_INTERACTIVE:-"bash nano tree"}
ARG APK_RUNTIME=${APK_RUNTIME:-"go git openssl ca-certificates"}
ARG APK_BUILD=${APK_BUILD:-"gcc g++ musl-dev gfortran lapack-dev openssl-dev"}

ENV APP_BASENAME=${APP_BASENAME:-"scraper"} \
    PATH="${GOPATH}/bin:/app:$PATH" \
    GOPATH=${GOPATH:-"/go"}

RUN \
      apk add --no-cache ${APK_RUNTIME} && \
    \
      apk add --no-cache --virtual=.interactive-dependencies ${APK_INTERACTIVE} && \
    \
      apk add --no-cache --virtual=.build-dependencies ${APK_BUILD} && \
    \
    mkdir -p /data/cache

#    \
#      apk del --no-cache --virtual=.build-dependencies && \

COPY . /go/src/github.com/roscopecoltran/scraper
WORKDIR /go/src/github.com/roscopecoltran/scraper

RUN \
    go get -v -u github.com/Masterminds/glide && \
    go get -v -u github.com/mitchellh/gox && \
    \
    glide install --strip-vendor

#    \
#    gox -verbose -os="linux" -arch="amd64" -output="/app/{{.Dir}}" ./cmd/scraper-server
# apk add --no-cache oniguruma-dev

VOLUME ["/data"]

EXPOSE 3000

CMD ["/bin/bash"]