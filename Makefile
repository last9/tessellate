protodep:
	go get -v github.com/golang/protobuf/protoc-gen-go
	go get -v github.com/lyft/protoc-gen-validate
	protoc --version || /bin/bash install_protobuf.sh

proto: protodep
	protoc \
		-I. \
		-I${GOPATH}/src \
		--go_out=plugins=grpc:${GOPATH}/src \
		--validate_out="lang=go:${GOPATH}/src" \
		proto/tessellate.proto

test: protodep
	docker-compose -f docker-compose.yaml up -d
	go.py go test -v -run TestStorer ./storage/...
	go.py go test -v ./runner/...
	docker-compose -f docker-compose.yaml stop