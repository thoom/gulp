docker run --rm -it -v $PWD:/go/src/gulp -w /go/src/gulp golang:alpine sh -c "apk add --update --no-cache git && go get -d ./... && go build -o gulp-alpine"
docker build -t gulp -f Dockerfile-alpine .

docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag gulp thoom/gulp:latest
docker tag gulp thoom/gulp:$TRAVIS_BRANCH

docker push thoom/gulp