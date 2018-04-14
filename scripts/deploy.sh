docker run --rm -it -v $PWD:/go/src/github.com/thoom/gulp -w /go/src/github.com/thoom/gulp -e TRAVIS_BRANCH golang:alpine sh -c "apk add --update --no-cache git && go get -d ./... && go build -ldflags \"-X github.com/thoom/gulp/client.buildVersion=$TRAVIS_BRANCH\" -o gulp-alpine"
docker build -t gulp -f Dockerfile-alpine --no-cache .

docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag gulp thoom/gulp:latest
docker tag gulp thoom/gulp:$TRAVIS_BRANCH

docker push thoom/gulp