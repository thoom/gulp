docker build -t gulp -f Dockerfile --no-cache --build-arg BUILD_VERSION=$TRAVIS_BRANCH .

docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag gulp thoom/gulp:latest
docker tag gulp thoom/gulp:$TRAVIS_BRANCH

docker push thoom/gulp