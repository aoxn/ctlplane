REPO?=registry.cn-hangzhou.aliyuncs.com
NAMESPACE?=spacexnice
APP_NAME?=hub-console
TAG?=1.0.0

IMAGE=${REPO}/${NAMESPACE}/${APP_NAME}:${TAG}


all:
	bash build.sh
	docker build -t ${IMAGE} build
	docker push ${IMAGE}

build:
	docker build -t ${IMAGE} build

test:
    echo test

push: build
	docker push ${IMAGE}

clean:
    rm -rf build

.PHONY: all