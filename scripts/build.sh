docker run --rm -it -v $PWD:/go/src/gulp -w /go/src/gulp golang:alpine sh -c "apk add --update --no-cache git && go get -d ./... && go build -o gulp-alpine"
env GOOS=linux GOARCH=386 go build -o gulp-linux-386
env GOOS=linux GOARCH=amd64 go build -o gulp-linux-amd64
env GOOS=darwin GOARCH=amd64 go build -o gulp-darwin