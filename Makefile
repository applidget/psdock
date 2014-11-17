# make .
# make -C path/dir


# Note: make release can't rely on cross compiling (https://code.google.com/p/go/issues/detail?id=6376)
#       must be ran on linux and darwin

HARDWARE=$(shell uname -m)
OS=$(shell uname -s)

test:
	go test

build:
	$(info ==> psdock binary will be in GOPATH/bin/psdock (tip: add GOPATH/bin to your PATH))
	go install ./cmd/psdock/
	
release:
	mkdir -p release
	cd ./cmd/psdock && go build -o ../../release/psdock
	cd release && tar -zcf psdock_$(OS)_$(HARDWARE).tar.gz psdock
	rm release/psdock
