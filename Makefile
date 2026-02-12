# Set these to the desired values
ARTIFACT_ID=ces-confd
VERSION=0.11.0

MAKEFILES_VERSION=10.6.0
GOTAG=1.25.7

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
include build/make/mocks.mk
include build/make/trivyscan.mk

