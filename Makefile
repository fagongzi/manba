RELEASE_VERSION    = $(release_version)

ifeq ("$(RELEASE_VERSION)","")
	RELEASE_VERSION		:= "unknown"
endif

ROOT_DIR 		   = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/
VERSION_PATH	   = $(shell echo $(ROOT_DIR) | sed -e "s;${GOPATH}/src/;;g")pkg/util
LD_GIT_COMMIT      = -X '$(VERSION_PATH).GitCommit=`git rev-parse --short HEAD`'
LD_BUILD_TIME      = -X '$(VERSION_PATH).BuildTime=`date +%FT%T%z`'
LD_GO_VERSION      = -X '$(VERSION_PATH).GoVersion=`go version`'
LD_MANBA_VERSION = -X '$(VERSION_PATH).Version=$(RELEASE_VERSION)'
LD_FLAGS           = -ldflags "$(LD_GIT_COMMIT) $(LD_BUILD_TIME) $(LD_GO_VERSION) $(LD_MANBA_VERSION) -w -s"

GOOS 		= linux
CGO_ENABLED = 0
DIST_DIR 	= $(ROOT_DIR)dist/

ETCD_VER			= v3.3.12
ETCD_DOWNLOAD_URL	= https://github.com/coreos/etcd/releases/download

MY_TARGET := dist_dir
EXEC_NAME := manba-proxy
IMAGE_NAME := manba
CMD_NAME := demo
ifeq ("$(MAKECMDGOALS)","docker")
	ifeq ("$(with)","")
		MY_TARGET := release download_etcd ui
	endif
	ifeq ($(findstring etcd,$(with)),etcd)
		MY_TARGET := $(MY_TARGET) download_etcd
		EXEC_NAME := etcd
		IMAGE_NAME = etcd
		CMD_NAME   = etcd
	endif
	ifeq ($(findstring apiserver,$(with)),apiserver)
		MY_TARGET := $(MY_TARGET) apiserver ui
		EXEC_NAME := apiserver
		IMAGE_NAME = apiserver
		CMD_NAME   = apiserver
	endif
	ifeq ($(findstring proxy,$(with)),proxy)
		MY_TARGET := $(MY_TARGET) proxy
		EXEC_NAME := manba-proxy
		IMAGE_NAME = proxy
		CMD_NAME   = manba-proxy
	endif
endif

.PHONY: release
release: dist_dir apiserver proxy;

.PHONY: release_darwin
release_darwin: darwin dist_dir apiserver proxy;

.PHONY: docker
docker:
	@$(MAKE) $(MY_TARGET)
	@echo ========== current docker tag is: $(RELEASE_VERSION) ==========
	docker build -t fagongzi/$(IMAGE_NAME):$(RELEASE_VERSION) \
				--build-arg EXEC_NAME="$(EXEC_NAME)" \
				--build-arg CMD_NAME="$(CMD_NAME)" \
				-f Dockerfile .
	docker tag fagongzi/$(IMAGE_NAME):$(RELEASE_VERSION) fagongzi/$(IMAGE_NAME)

.PHONY: ui
ui: ; $(info ======== compile ui:)
	git clone https://github.com/fagongzi/gateway-ui-vue.git $(DIST_DIR)ui
	cd $(DIST_DIR)ui && git checkout 3.0.0

.PHONY: darwin
darwin:
	$(eval GOOS := darwin)

.PHONY: apiserver
apiserver: ; $(info ======== compiled apiserver:)
	env CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -mod vendor -a -installsuffix cgo -o $(DIST_DIR)manba-apiserver $(LD_FLAGS) $(ROOT_DIR)cmd/api/*.go

.PHONY: proxy
proxy: ; $(info ======== compiled proxy:)
	env CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -mod vendor -a -installsuffix cgo -o $(DIST_DIR)manba-proxy $(LD_FLAGS) $(ROOT_DIR)cmd/proxy/*.go

.PHONY: download_etcd
download_etcd:
	curl -L $(ETCD_DOWNLOAD_URL)/$(ETCD_VER)/etcd-$(ETCD_VER)-linux-amd64.tar.gz -o /tmp/etcd-$(ETCD_VER)-linux-amd64.tar.gz
	tar xzvf /tmp/etcd-$(ETCD_VER)-linux-amd64.tar.gz -C $(DIST_DIR) --strip-components=1
	@rm -rf $(DIST_DIR)Documentation $(DIST_DIR)README*

.PHONY: dist_dir
dist_dir: ; $(info ======== prepare distribute dir:)
	mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)*
	@cp entrypoint.sh $(DIST_DIR)

.PHONY: clean
clean: ; $(info ======== clean all:)
	rm -rf $(DIST_DIR)*

.PHONY: help
help:
	@echo "build release binary: \n\t\tmake release\n"
	@echo "build Mac OS X release binary: \n\t\tmake release_darwin\n"
	@echo "build docker release with etcd: \n\t\tmake docker\n"
	@echo "\t  add 「with」 params can select what you need:\n"
	@echo "\t  default: all, like 「make docker」\n"
	@echo "\t  etcd: download and extract etcd and etcdctl\n"
	@echo "\t  proxy: only compile proxy\n"
	@echo "\t  apiserver: compile apiserver and download ui\n"
	@echo "clean all binary: \n\t\tmake clean\n"

UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)
	.DEFAULT_GOAL := release_darwin
else
	.DEFAULT_GOAL := release
endif
