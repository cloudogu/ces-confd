# Set these to the desired values
ARTIFACT_ID=ces-confd
VERSION=0.4.0

MAKEFILES_VERSION=4.3.0

.DEFAULT_GOAL:=compile

include build/make/variables.mk

include build/make/self-update.mk
include build/make/info.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-integration.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/package-tar.mk
