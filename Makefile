build:
	gom install
	gom test
	go build pki.io.go helpers.go runAPI.go runCA.go  runCert.go  runEntity.go  runOrg.go runAdmin.go runCSR.go runClient.go  runInit.go runNode.go runPairingKey.go

install:
	install -m 0755 pki.io /usr/local/bin
test:
	#export GOPATH=$(pwd)/../../
	bats bats
clean:
	rm pki.io

all: test build install
default: test build
