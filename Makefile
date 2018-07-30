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

ifeq ($(strip $(CONSUL_ADDR)),)
CONSUL_ADDR = "127.0.0.1:8500"
endif

test: build_deps start_server
	go test -v ./...
	make stop_server

http_build:
	protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:${GOPATH}/src \
		proto/tessellate.proto
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_http -a -installsuffix cgo \
		github.com/tsocial/tessellate/commands/http

worker_build: build_deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_worker -a -installsuffix cgo \
		github.com/tsocial/tessellate/commands/worker

tessellate_build: build_deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate -a -installsuffix \
		cgo github.com/tsocial/tessellate/

cli_build: build_deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate_cli -a -installsuffix \
		cgo github.com/tsocial/tessellate/commands/cli

tessellate: build_deps tessellate_build

start_server: tessellate
	nohup ./tessellate --support-version 0.0.4 >/dev/null &

stop_server:
	pkill tessellate

worker: build_deps worker_build
http: build_deps http_build

build_images: worker_build tessellate_build
	docker-compose -f docker-compose.yaml build worker
	docker-compose -f docker-compose.yaml build grpc-server

upload_worker: docker_login
	docker tag $(WORKER_REPO):latest $(WORKER_REPO):$(TRAVIS_BRANCH)-latest
	docker tag $(WORKER_REPO):latest $(WORKER_REPO):$(TRAVIS_BRANCH)-$(TRAVIS_BUILD_NUMBER)
	docker push $(WORKER_REPO):latest
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
	echo "$(DOCKER_PASSWORD)" | docker login registry.gitlab.com -u "$(DOCKER_USERNAME)" --password-stdin
