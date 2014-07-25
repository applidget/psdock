# make .
# make -C path/dir
 
build: psdock
	$(info ==> psdock binary in $(GOPATH)/bin/psdock (tip: add the $(GOPATH)/bin to your PATH))
 
psdock: lib_psdock
	godep go install ./cmd/psdock/
 
lib_psdock: check_env
 
check_env:
ifndef GOPATH
	$(error GOPATH must be set)
endif