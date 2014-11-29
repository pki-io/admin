# OSX makefile
default: all
all:
	go build pki.io.go runAPI.go
clean:
	rm pki.io
