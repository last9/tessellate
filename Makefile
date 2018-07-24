WORKER_REPO := "tsl8/worker"
SERVER_REPO := "tsl8/grpc-server"

.PHONY: worker tessellate http

protodep:
	go get -v github.com/golang/protobuf/protoc-gen-go
	go get -v github.com/lyft/protoc-gen-validate
	go get -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	protoc --version || /bin/bash install_protobuf.sh

deps:
	dep version || (curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh)
	dep ensure -v

clean:
	rm -rf vendor/github.com/hashicorp/nomad/nomad/structs/structs.generated.go

proto: protodep
	protoc \
		-I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:${GOPATH}/src \
		--validate_out="lang=go:${GOPATH}/src" \
		proto/tessellate.proto

build_deps: proto deps clean

test: build_deps
	go test -v ./...

http_build:
	protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:${GOPATH}/src \
		proto/tessellate.proto
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_http -a -installsuffix cgo \
		gitlab.com/tsocial/sre/tessellate/commands/http

worker_build:
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_worker -a -installsuffix cgo \
		gitlab.com/tsocial/sre/tessellate/commands/worker

tessellate_build:
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate -a -installsuffix \
		cgo gitlab.com/tsocial/sre/tessellate/

cli_build:
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_cli -a -installsuffix \
		cgo github.com/tsocial/sre/tessellate/commands/cli

tessellate: build_deps tessellate_build
worker: build_deps worker_build
http: build_deps http_build

build_images: worker_build tessellate_build
	docker-compose -f docker-compose.yaml build worker
	docker-compose -f docker-compose.yaml build grpc-server

upload_worker: docker_login
	docker tag $(WORKER_REPO):latest $(WORKER_REPO):$(TRAVIS_BRANCH)-latest
	docker tag $(WORKER_REPO):latest $(WORKER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)
	docker push $(SERVER_REPO):latest
	docker push $(WORKER_REPO):$(TRAVIS_BRANCH)-latest
	docker push $(WORKER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)

upload_server: docker_login
	docker tag $(SERVER_REPO):latest $(SERVER_REPO):$(TRAVIS_BRANCH)-latest
	docker tag $(SERVER_REPO):latest $(SERVER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)
	docker push $(SERVER_REPO):latest
	docker push $(SERVER_REPO):$(TRAVIS_BRANCH)-latest
	docker push $(SERVER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)

upload_images: upload_worker upload_server

docker_login:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
