VERSION ?= $(shell git describe --tags --always)

IMAGE = dkr.yldr.io/vulcand
PKG = github.com/yieldr/vulcand

OS ?= darwin
ARCH ?= amd64

GOBUILDFLAGS = -a -tags netgo -ldflags '-w'

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/vulcand $(GOBUILDFLAGS)
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/vctl $(GOBUILDFLAGS) ./vctl

install:
	go install .
	go install ./vctl

test:
	go test

docker-all: docker-build docker-image docker-push

docker-build:
	@docker run -i --rm -v "$(PWD):/go/src/$(PKG)" $(IMAGE):build make build

docker-test:
	@docker run -i --rm -v "$(PWD):/go/src/$(PKG)" $(IMAGE):build make test

docker-image:
	@docker build -t $(IMAGE):$(VERSION) .
	@docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

docker-push:
	@docker push $(IMAGE):$(VERSION)
	@docker push $(IMAGE):latest

docker-builder-image:
	@docker build -t $(IMAGE):build -f Dockerfile.build .
