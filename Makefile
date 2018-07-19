.PHONY: worker

protodep:
	go get -v github.com/golang/protobuf/protoc-gen-go
	go get -v github.com/lyft/protoc-gen-validate
	go get -v github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	protoc --version || /bin/bash install_protobuf.sh

deps:
	dep version || (curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh)
	dep ensure -v
	rm -rf vendor/github.com/hashicorp/nomad/nomad/structs/structs.generated.go

proto: protodep
	protoc \
		-I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=plugins=grpc:${GOPATH}/src \
		--validate_out="lang=go:${GOPATH}/src" \
		proto/tessellate.proto

http: proto deps
	protoc -I. \
		-I${GOPATH}/src \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--grpc-gateway_out=logtostderr=true:${GOPATH}/src \
		proto/tessellate.proto
	go build github.com/tsocial/tessellate/commands/http

test: proto deps
	go test -v ./...

worker: deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o worker -a -installsuffix cgo github.com/tsocial/tessellate/commands/worker

tessellate: deps
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o tessellate -a -installsuffix cgo github.com/tsocial/tessellate/

build_images: worker tessellate http
	docker-compose -f docker-compose.yaml build worker
	docker-compose -f docker-compose.yaml build tessellate
	docker-compose -f docker-compose.yaml build http

upload_images: build_images docker_login
	docker push worker
	docker push tessellate
	docker push http

docker_login:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
