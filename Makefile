all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goproxy cmd/proxy/main.go

.PHONY: docker-registry-login
docker-registry-login:
	$$(aws ecr get-login --no-include-email --region us-east-1 --profile=authereum)

.PHONY: docker-image-build
docker-image-build:
	docker build -t authereum/rpc-provider-proxy .

.PHONY: docker-image-run
docker-image-run:
	docker run -p 8000:8000 authereum/rpc-provider-proxy ./goproxy -proxy-url="https://kovan.authereum.com" -proxy-method=POST

.PHONY: docker-image-tag
docker-image-tag:
	$(eval REV=$(shell git rev-parse HEAD | cut -c1-7))
	docker tag authereum/rpc-provider-proxy:latest 874777227511.dkr.ecr.us-east-1.amazonaws.com/authereum/rpc-provider-proxy:latest
	docker tag authereum/rpc-provider-proxy:latest 874777227511.dkr.ecr.us-east-1.amazonaws.com/authereum/rpc-provider-proxy:$(REV)

.PHONY: docker-registry-push
docker-registry-push:
	$(eval REV=$(shell git rev-parse HEAD | cut -c1-7))
	docker push 874777227511.dkr.ecr.us-east-1.amazonaws.com/authereum/rpc-provider-proxy:latest
	docker push 874777227511.dkr.ecr.us-east-1.amazonaws.com/authereum/rpc-provider-proxy:$(REV)

.PHONY: docker-build-and-push
docker-build-and-push: docker-registry-login docker-image-build docker-image-tag docker-registry-push
