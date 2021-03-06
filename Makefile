.DEFAULT_GOAL=all
PACKAGES_WITH_TESTS:=$(shell go list -f="{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}" ./... | grep -v '/vendor/')
TEST_TARGETS:=$(foreach p,${PACKAGES_WITH_TESTS},test-$(p))
TEST_OUT_DIR:=testout

SCALA_PB  := github.com/scalapb/ScalaPB
PROTO_SRC := src/main/protobuf

PROJECT:= $(subst ${GOPATH}/src/,,$(shell pwd))

DOCKER_TARGET_REGISTRY        ?=
BUILD_NUMBER                  ?= 1
GIT_LAST_COMMIT_ID            ?= $(shell git rev-parse --short HEAD)
GIT_CURRENT_BRANCH            ?= $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_TARGET_IMAGE_TAG       ?= $(BUILD_NUMBER)-$(subst /,-,$(GIT_CURRENT_BRANCH))-$(GIT_LAST_COMMIT_ID)
DOCKER_TARGET_IMAGE_NAME      ?= $(shell basename $(shell git rev-parse --show-toplevel))
DOCKER_TARGET_IMAGE           ?= $(DOCKER_TARGET_REGISTRY)$(DOCKER_TARGET_IMAGE_NAME):$(DOCKER_TARGET_IMAGE_TAG)
DOCKER_BUILD_WORKSPACE_SUBDIR ?=
DOCKER_BUILD_WORKSPACE_DIR    ?= $(shell realpath $(shell git rev-parse --show-toplevel)/$(DOCKER_BUILD_WORKSPACE_SUBDIR))
MICROSERVICE                  ?= $(DOCKER_TARGET_IMAGE_NAME)
ARTIFACT_TYPE                 ?= tar
DOCKER_FILE                   ?= Dockerfile

.PHONY: all
all: gencode proto-golang proto-py mod lint testall docker-build

.PHONY: mod
mod:
	rm -rf vendor
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod download

.PHONY: lint
lint:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-linter:latest \
	golangci-lint run ./... --exclude "is deprecated"

.PHONY: gencode
gencode:
	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-easyjson:latest \
	sh -c 'rm -fv pkg/provider/*/*_easyjson.go && \
	easyjson -all pkg/provider/fcm/request.go && \
	easyjson -all pkg/provider/fcm/response.go && \
	easyjson -all pkg/provider/ans/request.go && \
	easyjson -all pkg/provider/ans/response.go && \
	easyjson -all pkg/provider/gcm/request.go && \
	easyjson -all pkg/provider/gcm/response.go'

.PHONY: proto-golang
proto-golang: protoc-scalapb
	$(eval $@_target :=pkg/api)

	rm -f ${$@_target}/*.pb.go

	docker run -it --rm \
	-v "$(shell pwd):/go/src/${PROJECT}" \
	-v "${GOPATH}/pkg:/go/pkg" \
	-w "/go/src/${PROJECT}" \
	-e "GOFLAGS=" \
	go-tools-protoc:latest \
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

.PHONY: protoc-scalapb
protoc-scalapb:
	$(eval $@_target :=vendor/${SCALA_PB})
	rm -rf ${$@_target}
	mkdir -m 755 -p ${$@_target}
	git clone -b master https://${SCALA_PB} ${$@_target}

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
docker-build: docker-clean
	docker build -f ${DOCKER_FILE} \
	--tag ${DOCKER_TARGET_IMAGE} \
	--build-arg "COMMIT=${GIT_LAST_COMMIT_ID}" \
	--build-arg "RELEASE=${DOCKER_TARGET_IMAGE_TAG}" \
	.

.PHONY: docker-image-save
docker-image-save:
	docker image save $(DOCKER_TARGET_IMAGE) | gzip > $(MICROSERVICE).$(ARTIFACT_TYPE)

.PHONY: docker-push
docker-push:
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASSWORD) $(DOCKER_TARGET_REGISTRY)
	docker push $(DOCKER_TARGET_IMAGE)

.PHONY: docker-clean
docker-clean:
	-docker rm -f $(docker ps -a -q --filter=ancestor=${DOCKER_TARGET_IMAGE})
	-docker rmi -f $(docker images -q ${DOCKER_TARGET_IMAGE})
	-docker rmi $(docker images -f "dangling=true" -q)

.PHONY: docker-run
docker-run:
	docker run --rm -it \
	-p "8010:8010" \
	-p "8011:8011" \
	-v "$(shell pwd)/example.yaml:/var/config/example.yaml" \
	-v "${HOME}/<...>.pem:/config/production-big.pem" \
	-v "${HOME}/<...>.pem:/config/production-big-voip.pem" \
	-v "${HOME}/<...>.pem:/config/production-ee.pem" \
	-v "${HOME}/<...>.pem:/config/production-ee-voip.pem" \
	${DOCKER_TARGET_IMAGE} \
	sh -c "push-server -c /var/config/example.yaml"

.PHONY: scala-publish-local
scala-publish-local:
	sbt clean compile +publishLocal
