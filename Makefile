SHELL := /bin/bash

# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"
# hey -m GET -c 100 -n 10000  http://localhost:3000/v1/test

# To generate a private/public key PEM file.
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem

# ==============================================================================
# Building containers

VERSION := 1.0

all: docker-build docker-push

docker-build:
	docker build \
		-t b65b0111-kr1-registry.container.cloud.toast.com/dns-controller:$(version) \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse --short HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.
# ==============================================================================

docker-push:
	docker push b65b0111-kr1-registry.container.cloud.toast.com/dns-controller:$(version)

tidy:
	go mod tidy
	go mod vendor

run:
	go run main.go

test:
	go test -v ./... -count=1
	#staticcheck ./...

push:
	git add -A
	git commit -m "update"
	git push origin master

pull:
	git pull origin master