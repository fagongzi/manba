ROOT_DIR 		= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/
VERSION_PATH	= $(shell echo $(ROOT_DIR) | sed -e "s;${GOPATH}/src/;;g")pkg/util
LD_GIT_COMMIT   = -X '$(VERSION_PATH).GitCommit=`git rev-parse --short HEAD`'
LD_BUILD_TIME   = -X '$(VERSION_PATH).BuildTime=`date +%FT%T%z`'
LD_GO_VERSION   = -X '$(VERSION_PATH).GoVersion=`go version`'
LD_FLAGS        = -ldflags "$(LD_GIT_COMMIT) $(LD_BUILD_TIME) $(LD_GO_VERSION) -w -s"

GOOS 		= linux
CGO_ENABLED = 0
DIST_DIR 	= $(ROOT_DIR)dist/

ETCD_VER			= v3.0.14
ETCD_DOWNLOAD_URL	= https://github.com/coreos/etcd/releases/download

DOCKER_TAG			:= $(tag)

ifeq ("$(DOCKER_TAG)","")
	DOCKER_TAG		:= $(shell date +%Y%m%d%H%M)
endif

.PHONY: release
release: dist_dir apiserver proxy;

.PHONY: release_darwin
release_darwin: darwin dist_dir apiserver proxy;

.PHONY: docker
docker: release download_etcd;
	@echo ========== current docker tag is: $(DOCKER_TAG) ==========
	docker build -t fagongzi/gateway:$(DOCKER_TAG) -f Dockerfile .
	docker build -t fagongzi/proxy:$(DOCKER_TAG) -f Dockerfile-proxy .
	docker build -t fagongzi/apiserver:$(DOCKER_TAG) -f Dockerfile-apiserver .

.PHONY: darwin
darwin:
	$(eval GOOS := darwin)

.PHONY: apiserver
apiserver: ; $(info ======== compiled apiserver:)
	env CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o $(DIST_DIR)apiserver $(LD_FLAGS) $(ROOT_DIR)cmd/api/*.go

.PHONY: proxy
proxy: ; $(info ======== compiled proxy:)
	env CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -a -installsuffix cgo -o $(DIST_DIR)proxy $(LD_FLAGS) $(ROOT_DIR)cmd/proxy/*.go

.PHONY: download_etcd
download_etcd:
	curl -L $(ETCD_DOWNLOAD_URL)/$(ETCD_VER)/etcd-$(ETCD_VER)-linux-amd64.tar.gz -o /tmp/etcd-$(ETCD_VER)-linux-amd64.tar.gz
	tar xzvf /tmp/etcd-$(ETCD_VER)-linux-amd64.tar.gz -C $(DIST_DIR) --strip-components=1
	@rm -rf $(DIST_DIR)Documentation $(DIST_DIR)README*

.PHONY: dist_dir
dist_dir: ; $(info ======== prepare distribute dir:)
	mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)*

.PHONY: clean
clean: ; $(info ======== clean all:)
	rm -rf $(DIST_DIR)*

.PHONY: help
help:
	@echo "build release binary: \n\t\tmake release\n"
	@echo "build Mac OS X release binary: \n\t\tmake release_darwin\n"
	@echo "build docker release with etcd: \n\t\tmake docker\n"
	@echo "clean all binary: \n\t\tmake clean\n"

UNAME_S := $(shell uname -s)

# 设置默认编译目标
ifeq ($(UNAME_S),Darwin)
	.DEFAULT_GOAL := release_darwin
else
	.DEFAULT_GOAL := release
endif
