env GOOS=linux GOARCH=386 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION-linux386" -o gulp-linux-386
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION-linux64" -o gulp-linux-amd64
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION-darwin-amd64" -o gulp-darwin-amd64
env GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION-darwin-arm64" -o gulp-darwin-arm64
env GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION-windows" -o gulp-windows