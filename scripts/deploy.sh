docker build -t gulp -f Dockerfile-alpine .

docker login -u $DOCKER_USER -p $DOCKER_PASS
docker tag gulp thoom/gulp:latest
docker tag gulp thoom/gulp:$TRAVIS_BUILD_NUMBER

docker push thoom/gulp