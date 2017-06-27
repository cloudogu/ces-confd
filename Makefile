#
# useful targets:
# 
# update-dependencies
#	calls glide to recreate glide.lock and update dependencies
#
# info
#	prints build time information
#
# build
#	builds the executable and ubuntu packages for trusty and xenial
#
# unit-test
#	performs unit-testing
#
# integration-test
#	not implemented yet
#
# static-analysis
#	performs static source code analysis
#
# deploy
#	deploys ubuntu packages for trusty and xenial to the according repositories
#
# undeploy
#	undeploys ubuntu packages for trusty and xenial from the according repositories
#
# clean
#	remove target folder
#
# dist-clean
#	also removes any downloaded dependencies
#

# collect packages and dependencies for later usage
PACKAGES=$(shell go list ./... | grep -v /vendor/)

ARTIFACT_ID=ces-confd
VERSION=0.1.5
BUILD_TIME:=$(shell date +%FT%T%z)
COMMIT_ID:=$(shell git rev-parse HEAD)


# if VERSION is not set, try to guess from git. Experimental
#
ifndef VERSION
BRANCH:=$(shell git symbolic-ref HEAD >/dev/null 2>&1 && git symbolic-ref HEAD | sed s@"^refs/heads/"@@ || git rev-parse --short HEAD 2>/dev/null |sed s@"^refs/heads/"@@)
BRANCH_TYPE=$(patsubst master,master,$(patsubst develop,develop,$(patsubst feature/%,feature,$(patsubst release/%,release,$(patsubst hotfix/%,hotfix,$(BRANCH))))))
VERSION_MAJOR=$(shell git describe |sed s/"^v\([0-9][0-9]*\).[0-9][0-9]*.[0-9][0-9]*.*"/"\1"/)
VERSION_MINOR=$(shell git describe |sed s/"^v[0-9][0-9]*.\([0-9][0-9]*\).[0-9][0-9]*.*"/"\1"/)
VERSION_PATCH=$(shell git describe |sed s/"^v[0-9][0-9]*.[0-9][0-9]*.\([0-9][0-9]*\).*"/"\1"/)
EXACT_MATCH=$(shell git describe --exact-match >/dev/null 2>&1 || echo "yes")

  ifeq ('master',${BRANCH_TYPE})
    ifeq ('yes',${EXACT_MATCH})
    else
$(error untagged commit to master)
    endif
  else
SNAPSHOT=-SNAPSHOT
VERSION_PATCH:=$(shell echo "${VERSION_PATCH}+1" |bc -l)
  endif
VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH}${SNAPSHOT}
endif



# directory settings
TARGET_DIR=target/
COMPILE_TARGET_DIR=target/dist/

# make target files
EXECUTABLE=target/dist/${ARTIFACT_ID}
PACKAGE=target/dist/${ARTIFACT_ID}-${VERSION}.tar.gz
XUNIT_XML=target/unit-tests.xml

# tools
LINT=gometalinter
GLIDE=glide
GO2XUNIT=go2xunit

# flags
LINTFLAGS=--vendor --exclude="vendor" --exclude="_test.go"
LINTFLAGS+=--disable-all --enable=errcheck --enable=vet --enable=golint
LINTFLAGS+=--deadline=2m
LDFLAGS=-ldflags "-extldflags -static -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitID=${COMMIT_ID}"
GLIDEFLAGS=


# choose the environment, if BUILD_URL environment variable is available then we are on ci (jenkins)
ifdef BUILD_URL
ENVIRONMENT=ci
GLIDEFLAGS+=--no-color
else
ENVIRONMENT=local
endif


# default goal is "build"
#
.DEFAULT_GOAL:=build


# updating dependencies
#
update-dependencies: glide.lock

glide.lock: glide.yaml
	${GLIDE} ${GLIDEFLAGS} up -v


# build steps: dependencies, compile, package
#
# XXX dependencies- target can not be associated to a file.
# As a consequence make build will always trigger a full build, even if targets already exist.
#
info:
	@echo "dumping build information ..."
	@echo "Version    : $(VERSION)"
	@echo "Snapshot   : $(SNAPSHOT)"
	@echo "Build-Time : $(BUILD_TIME)"
	@echo "Commit-ID  : $(COMMIT_ID)"
	@echo "Environment: $(ENVIRONMENT)"
	@echo "Branch     : $(BRANCH)"
	@echo "Branch-Type: $(BRANCH_TYPE)"
	@echo "Packages   : $(PACKAGES)"

dependencies: info
	@echo "installing dependencies ..."
	${GLIDE} ${GLIDEFLAGS} install -v

${EXECUTABLE}: dependencies
	@echo "compiling ..."
	mkdir -p $(COMPILE_TARGET_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo ${LDFLAGS} -o $@
	@echo "... executable can be found at $@"

${PACKAGE}: ${EXECUTABLE}
	cd ${COMPILE_TARGET_DIR} && tar cvzf ${ARTIFACT_ID}-${VERSION}.tar.gz ${ARTIFACT_ID}

build: ${PACKAGE}


# unit tests
#
unit-test: ${XUNIT_XML}

${XUNIT_XML}:
	mkdir -p $(TARGET_DIR)
	go test -v $(PACKAGES) | tee target/unit-tests.log
	@if grep '^FAIL' target/unit-tests.log; then \
		exit 1; \
	fi
	@if grep '^=== RUN' target/unit-tests.log; then \
	  cat target/unit-tests.log | go2xunit -output $@; \
	fi


# integration tests, not yet
#
integration-test:
	@echo "not yet implemented"


# static analysis
#
static-analysis: static-analysis-${ENVIRONMENT}

static-analysis-ci: target/static-analysis-cs.log
	@if [ X"$${CI_PULL_REQUEST}" != X"" -a X"$${CI_PULL_REQUEST}" != X"null" ] ; then cat $< | CI_COMMIT=$(COMMIT_ID) reviewdog -f=checkstyle -ci="common" ; fi

static-analysis-local: target/static-analysis-cs.log target/static-analysis.log
	@echo ""
	@echo "differences to develop branch:"
	@echo ""
	@cat $< | reviewdog -f checkstyle -diff "git diff develop"

target/static-analysis.log:
	@mkdir -p ${TARGET_DIR}
	@echo ""
	@echo "complete static analysis:"
	@echo ""
	@$(LINT) ${LINTFLAGS} ./... | tee $@

target/static-analysis-cs.log:
	@mkdir -p ${TARGET_DIR}
	@$(LINT) ${LINTFLAGS} --checkstyle ./... > $@ | true


# clean lifecycle
#
clean:
	rm -rf ${TARGET_DIR}

dist-clean: clean
	rm -rf node_modules
	rm -rf public/vendor
	rm -rf vendor
	rm -rf npm-cache
	rm -rf bower

.PHONY: update-dependencies
.PHONY: build dependencies info
.PHONY: static-analysis static-analysis-ci static-analysis-local
.PHONY: integration-test
.PHONY: unit-test
.PHONY: clean dist-clean
