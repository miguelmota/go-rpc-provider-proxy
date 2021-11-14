INFURA_ID = "84842078b09946638c03157f83405213"

all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goproxy cmd/proxy/main.go

.PHONY: docker-registry-login
docker-registry-login:
	docker hub login

.PHONY: docker-image-build
docker-image-build:
	docker build -t miguelmota/rpc-provider-proxy .

.PHONY: docker-image-run
docker-image-run:
	docker run -p 8000:8000 miguelmota/rpc-provider-proxy ./goproxy -proxy-url="https://kovan.infura.io/v3/$(INFURA_ID)" -proxy-method=POST -limit-per-second=10

.PHONY: docker-image-tag
docker-image-tag:
	$(eval REV=$(shell git rev-parse HEAD | cut -c1-7))
	docker tag miguelmota/rpc-provider-proxy:latest miguelmota/rpc-provider-proxy:$(REV)

.PHONY: docker-registry-push
docker-registry-push:
	$(eval REV=$(shell git rev-parse HEAD | cut -c1-7))
	docker push miguelmota/rpc-provider-proxy:latest
	docker push miguelmota/rpc-provider-proxy:$(REV)

.PHONY: docker-build-and-push
docker-build-and-push: docker-registry-login docker-image-build docker-image-tag docker-registry-push

.PHONY: loadtest
loadtest:
	./scripts/bench.sh

