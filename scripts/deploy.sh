docker build -t gulp -f Dockerfile --no-cache --build-arg BUILD_VERSION=$RELEASE_VERSION .

echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin
docker tag gulp $IMAGE_NAME:latest
docker tag gulp $IMAGE_NAME:$RELEASE_VERSION

docker push $IMAGE_NAME:latest
docker push $IMAGE_NAME:$RELEASE_VERSION

echo $GH_PASS | docker login ghcr.io -u $GH_USER --password-stdin
docker tag gulp ghcr.io/$IMAGE_NAME:latest
docker tag gulp ghcr.io/$IMAGE_NAME:$RELEASE_VERSION

docker push ghcr.io/$IMAGE_NAME:latest
docker push ghcr.io/$IMAGE_NAME:$RELEASE_VERSION