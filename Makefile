CORE_DIRS = api config crypto document entity fs index node x509

default: get-deps build test

get-deps:
	fdm

build:
	fdm build -o pki.io

install:
	install -m 0755 pki.io /usr/local/bin

test:
	bats bats_tests

clean:
	test ! -d _vendor || rm -rf _vendor/*
	test ! -e pki.io || rm pki.io

dev: clean
	FDM_ENV=DEV fdm --dev
	rm -rf _vendor/src/github.com/pki-io/core/* && \
	for d in $(CORE_DIRS); do (cd _vendor/src/github.com/pki-io/core && ln -s ../../../../../../core/$$d .); done && \
	rm -rf _vendor/src/github.com/pki-io/controller && \
	(cd _vendor/src/github.com/pki-io && ln -s ../../../../../controller .) && \
	rm -rf _vendor/pkg
	fdm --dev

all: get-deps build test install
