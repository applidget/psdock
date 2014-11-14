# make .
# make -C path/dir

test:
	go test

build:
	go install ./cmd/psdock/
	go install ./cmd/psdock-init/
	$(info ==> psdock binary in GOPATH/bin/psdock (tip: add the $(GOPATH)/bin to your PATH))

