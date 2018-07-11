FROM golang:alpine as goalpine

ARG BUILD_VERSION=snapshot
COPY . /go/src/github.com/thoom/gulp

WORKDIR /go/src/github.com/thoom/gulp
RUN apk add --update --no-cache git ca-certificates \
    && go get -d ./... \ 
    && CGO_ENABLED=0 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$BUILD_VERSION-docker" -o gulp

FROM scratch
LABEL author="Zach Peacock <zdp@thoomtech.com>"

COPY --from=goalpine /go/src/github.com/thoom/gulp/gulp /bin/
COPY --from=goalpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /gulp
ENTRYPOINT ["/bin/gulp"]