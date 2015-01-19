build:
	gom build pki.io.go helpers.go runAPI.go runCA.go  runCert.go  runEntity.go  runOrg.go runAdmin.go runCSR.go runClient.go  runInit.go runNode.go runPairingKey.go

install:
	gom exec install -m 0755 pki.io /usr/local/bin
test:
	gom exec bats bats_tests
clean:
	gom exec rm pki.io

all: test build install
default: test build
