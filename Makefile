default: get-deps build test

get-deps:
	gom install
build:
	gom build pki.io.go helpers.go runAPI.go runCA.go  runCert.go  runEntity.go  runOrg.go runAdmin.go runCSR.go runClient.go  runInit.go runNode.go runPairingKey.go

install:
	install -m 0755 pki.io /usr/local/bin
test:
	gom exec bats bats_tests
clean:
	rm pki.io

all: get-deps build test install
