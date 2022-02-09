docker build -t gulp -f Dockerfile --no-cache --build-arg BUILD_VERSION=$RELEASE_VERSION .

docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag gulp thoom/gulp:latest
docker tag gulp thoom/gulp:$RELEASE_VERSION

docker push thoom/gulp