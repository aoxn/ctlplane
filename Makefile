REPO=61.160.36.122:8080
PROJECT=sigma
APP=registry-console
VERSION=1.0.0

IMAGE=${REPO}/${PROJECT}/${APP}:${VERSION}


all:
    source build.sh
	docker build -t ${IMAGE} .
	docker push ${IMAGE}

.PHONY: all