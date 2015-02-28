DIRS = config crypto document entity fs index node x509

default: get-deps build test

get-deps:
	gom install
build:
	gom build pki.io.go adminApp.go runNode.go helpers.go runCA.go runOrg.go nodeApp.go runInit.go runPairingKey.go

install:
	install -m 0755 pki.io /usr/local/bin
test:
	bats bats_tests
clean:
	test ! -d _vendor || rm -rf _vendor/*
	test ! -e pki.io || rm pki.io

dev: clean get-deps
	test -d _vendor/src/github.com/pki-io/pki.io  && \
	rm -rf _vendor/src/github.com/pki-io/pki.io/* && \
	for d in $(DIRS); do (cd _vendor/src/github.com/pki-io/pki.io && ln -s ../../../../../../pki.io/$$d .); done && \
	rm -rf _vendor/pkg

all: get-deps build test install
