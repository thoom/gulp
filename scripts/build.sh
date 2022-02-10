env GOOS=linux GOARCH=386 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION" -o gulp && tar cfz gulp.linux-386.tar.gz gulp
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION" -o gulp && tar cfz gulp.linux-amd64.tar.gz gulp
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION" -o gulp && tar cfz gulp.darwin-amd64.tar.gz gulp
env GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION" -o gulp && tar cfz gulp.darwin-arm64.tar.gz gulp
env GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$RELEASE_VERSION" -o gulp && zip gulp.windows.zip gulp