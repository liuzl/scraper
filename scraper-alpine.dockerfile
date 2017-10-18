FROM alpine:edge
MAINTAINER Rosco Pecoltran <https://github.com/roscopecoltran>

# build: docker build -t scraper:alpine -f scraper-alpine.dockerfile --no-cache .
# run: docker run --rm -ti -p 3000:3000 -v `pwd`:/app scraper:alpine

ARG GOPATH=${GOPATH:-"/go"}
ARG APK_INTERACTIVE=${APK_INTERACTIVE:-"bash nano tree"}
ARG APK_RUNTIME=${APK_RUNTIME:-"go git openssl ca-certificates"}
ARG APK_BUILD=${APK_BUILD:-"gcc g++ musl-dev gfortran lapack-dev openssl-dev oniguruma-dev"}

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
    go get github.com/Masterminds/glide && \
    go get github.com/mitchellh/gox && \
    \
    go get golang.org/x/net/... && \
    go get github.com/qor/session && \
    go get github.com/qor/action_bar && \
    go get github.com/qor/help && \
    go get github.com/qor/qor && \
    go get github.com/qor/admin && \
    go get github.com/qor/serializable_meta && \
    go get github.com/qor/worker && \
    go get github.com/qor/sorting && \
    go get github.com/qor/roles && \
    go get github.com/qor/publish && \
    go get github.com/qor/publish2 && \
    go get github.com/qor/oss/... && \
    go get github.com/jinzhu/gorm/... && \
    go get github.com/go-sql-driver/mysql && \
    go get github.com/roscopecoltran/admin && \
    \
    glide install --strip-vendor

    # gox -verbose -os="linux" -arch="amd64" -output="/app/{{.Dir}}" ./cmd/scraper-server

VOLUME ["/data"]

EXPOSE 3000 4000

CMD ["/bin/bash"]
# CMD ["/app/scraper-server","/app/conf.d/providers.list.json"]