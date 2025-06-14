FROM golang:1.24-alpine as build

ARG BUILD_VERSION=snapshot
COPY . /thoom/gulp

WORKDIR /thoom/gulp
RUN apk add --update --no-cache git ca-certificates
RUN go get -d ./... 
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$BUILD_VERSION" -o gulp
RUN touch /tmp/hosts

FROM scratch
LABEL author="Zach Peacock <zach@thoom.net>"
LABEL org.opencontainers.image.source="https://github.com/thoom/gulp"

COPY --from=build /thoom/gulp/gulp /bin/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /tmp/hosts /etc/hosts

WORKDIR /gulp
ENTRYPOINT ["/bin/gulp"]