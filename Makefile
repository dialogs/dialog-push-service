.DEFAULT_GOAL=all
PACKAGES_WITH_TESTS:=$(shell go list -f="{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}" ./... | grep -v '/vendor/')
TEST_TARGETS:=$(foreach p,${PACKAGES_WITH_TESTS},test-$(p))
TEST_OUT_DIR:=testout

TAG            := 0.0.1
PROJECT        := dialog-push-service
COMMIT         := $(shell git rev-parse --short HEAD)
DOCKER_REGISTRY?=
IMAGE          := ${DOCKER_REGISTRY}${PROJECT}:${TAG}
SCALA_PB       := github.com/scalapb/ScalaPB
PROTO_SRC      := src/main/protobuf


.PHONY: all
all: gencode mod proto-golang proto-py lint testall docker-build

.PHONY: mod
mod:
	rm -rf vendor
	GO111MODULE=on go mod download
	GO111MODULE=on go mod vendor

	$(eval $@_target :=vendor/${SCALA_PB})
	rm -rf ${$@_target}
	git clone -b master https://${SCALA_PB} ${$@_target}

.PHONY: lint
lint:
ifeq ($(shell command -v golangci-lint 2> /dev/null),)
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.0
endif
	golangci-lint run ./... --exclude "is deprecated"

.PHONY: gencode
gencode:
ifeq ($(shell command -v easyjson 2> /dev/null),)
	go get -u github.com/mailru/easyjson/...
endif

	$(eval $@_target := pkg/provider/fcm)
	rm -f ${$@_target}/*_easyjson.go
	easyjson -all ${$@_target}/request.go
	easyjson -all ${$@_target}/response.go

	$(eval $@_target := pkg/provider/ans)
	rm -f ${$@_target}/*_easyjson.go
	easyjson -all ${$@_target}/request.go
	easyjson -all ${$@_target}/response.go

.PHONY: proto-golang
proto-golang:
	$(eval $@_target :=pkg/api)

	rm -f ${$@_target}/*.pb.go

	protoc \
	-I=${PROTO_SRC} \
	-I=vendor/${SCALA_PB}/protobuf \
	--gogoslick_out=\
	plugins=grpc,\
	Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
	Mscalapb/scalapb.proto=github.com/gogo/protobuf/types:\
	${$@_target} ${PROTO_SRC}/*.proto

.PHONY: proto-py
proto-py:
	$(eval $@_target :=python/push)

	-rm -rf ${$@_target}
	mkdir -p -m 755 ${$@_target}

	python2.7 \
	-m grpc_tools.protoc \
	-I=${PROTO_SRC} \
	-I=vendor/${SCALA_PB}/protobuf \
	--python_out=${$@_target} \
	--grpc_python_out=${$@_target} \
	${PROTO_SRC}/*.proto

.PHONY: testall
testall:
	rm -rf ${TEST_OUT_DIR}
	mkdir -p -m 755 $(TEST_OUT_DIR)
	$(MAKE) -j 1 $(TEST_TARGETS)
	@echo "=== tests: ok ==="

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	$(eval $@_package := $(subst test-,,$@))
	$(eval $@_filename := $(subst /,_,$($@_package)))

	@echo "== test directory $($@_package) =="

	@go test $($@_package) -v -race -coverprofile $(TEST_OUT_DIR)/$($@_filename)_cover.out \
    >> $(TEST_OUT_DIR)/$($@_filename).out \
   || ( echo 'fail $($@_package)' && cat $(TEST_OUT_DIR)/$($@_filename).out; exit 1);


.PHONY: docker-build
docker-build:
	-docker rm -f `docker ps -a -q --filter=ancestor=${IMAGE}`
	-docker rmi -f `docker images -q ${IMAGE}`
	-docker rmi $(docker images -f "dangling=true" -q)

	docker build -f Dockerfile \
	--tag ${IMAGE} \
	--build-arg "COMMIT=${COMMIT}" \
	--build-arg "RELEASE=${TAG}" \
	.

.PHONY: docker-run
docker-run:
	docker run --rm -it \
	-p "8010:8010" \
	-p "8011:8011" \
	-v "$(shell pwd)/example.yaml:/var/config/example.yaml" \
	${IMAGE} \
	sh -c "/push-server -c /var/config/example.yaml"

.PHONY: scala-publish-local
scala-publish-local:
	sbt clean compile publish-local