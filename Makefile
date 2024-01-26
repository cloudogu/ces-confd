# Set these to the desired values
ARTIFACT_ID=ces-confd
VERSION=0.9.0

MAKEFILES_VERSION=7.0.1
GOTAG=1.17.8

.DEFAULT_GOAL:=compile

include build/make/variables.mk

include build/make/self-update.mk
# include build/make/info.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-integration.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/release.mk
include build/make/package-tar.mk
include build/make/digital-signature.mk
