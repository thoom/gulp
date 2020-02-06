FROM golang:1.13-alpine as build

ARG BUILD_VERSION=snapshot
COPY . /thoom/gulp

WORKDIR /thoom/gulp
RUN apk add --update --no-cache git ca-certificates \
    && go get -d ./... \ 
    && CGO_ENABLED=0 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$BUILD_VERSION-docker" -o gulp

FROM scratch
LABEL author="Zach Peacock <zdp@thoomtech.com>"

COPY --from=build /thoom/gulp/gulp /bin/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /gulp
ENTRYPOINT ["/bin/gulp"]