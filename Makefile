DIRS = admin ca certificate csr node org pairingkey

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
	mkdir -p _vendor/src/github.com/pki-io/controllers  && \
	rm -rf _vendor/src/github.com/pki-io/controllers/* && \
	for d in $(DIRS); do (cd _vendor/src/github.com/pki-io/core && ln -s ../../../../../../core/$$d .); done && \
	rm -rf _vendor/pkg
	fdm --dev

all: get-deps build test install
