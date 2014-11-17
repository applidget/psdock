# make .
# make -C path/dir

HARDWARE=$(shell uname -m)

test:
	go test

build:
	go install ./cmd/psdock/
  $(info ==> psdock binary in GOPATH/bin/psdock (tip: add GOPATH/bin to your PATH))

release:
	mkdir -p release
	GOOS=linux go build -o release/psdock
	cd release && tar -zcf psdock_linux_$(HARDWARE).tgz psdock
	GOOS=darwin go build -o release/psdock
	cd release && tar -zcf psdock_darwin_$(HARDWARE).tgz psdock
	rm release/psdock
