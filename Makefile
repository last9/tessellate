.PHONY: worker

protodep:
	go get -v github.com/golang/protobuf/protoc-gen-go
	go get -v github.com/lyft/protoc-gen-validate
	protoc --version || /bin/bash install_protobuf.sh

deps:
	dep version || (curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh)
	dep ensure

proto: protodep
	protoc \
		-I. \
		-I${GOPATH}/src \
		--go_out=plugins=grpc:${GOPATH}/src \
		--validate_out="lang=go:${GOPATH}/src" \
		proto/tessellate.proto

test: protodep deps
	go test -v ./...

worker:
	env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o worker -a -installsuffix cgo github.com/tsocial/tessellate/commands/worker
	docker-compose -f docker-compose.yaml build worker
