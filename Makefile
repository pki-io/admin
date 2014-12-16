# OSX makefile
default: all
all:
	go build pki.io.go helpers.go runAPI.go runCA.go  runCert.go  runEntity.go  runOrg.go runAdmin.go runCSR.go runClient.go  runInit.go
install: all
    install -m 0755 pki.io /usr/local/bin
        
clean:
	rm pki.io
