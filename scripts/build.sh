env GOOS=linux GOARCH=386 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$TRAVIS_BRANCH-linux386" -o gulp-linux-386
env GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$TRAVIS_BRANCH-linux64" -o gulp-linux-amd64
env GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/thoom/gulp/client.buildVersion=$TRAVIS_BRANCH-darwin" -o gulp-darwin