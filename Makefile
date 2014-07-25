# make .
# make -C path/dir
 
build: psdock
	$(info ==> psdock binary in $(GOPATH)/bin/psdock (tip: add the $(GOPATH)/bin to your PATH))
 
psdock: check_env
	godep go install ./cmd/psdock/
 

check_env:
ifndef GOPATH
	$(error GOPATH must be set)
endif
