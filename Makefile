# make .
# make -C path/dir

build: psdock
	$(info ==> psdock binary in $(GOPATH)/bin/psdock (tip: add the $(GOPATH)/bin to your PATH))
	godep go test

psdock: check_env
	godep go build
	godep go install ./cmd/psdock/
	godep go install ./cmd/psdock-init/

check_env:
ifndef GOPATH
	$(error GOPATH must be set)
endif
