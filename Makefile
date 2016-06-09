NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' main.go)
TESTDIRS = $(shell find . -name '*_test.go' | cut -d/ -f1-2 | uniq | grep -v vendor)

converge: main.go cmd/* load/* resource/* vendor/**/*
	go build .

test: converge samples/*.hcl samples/errors/*.hcl
	go test -v ${TESTDIRS}
	./converge validate samples/*.hcl
	./converge fmt --check samples/*.hcl

samples/errors/*.hcl: converge
	@echo === validating $@ should fail ==
	./converge validate $@ || exit 0 && exit 1

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@

vendor: main.go cmd/* load/* resource/* exec/*
	godep save -t ./...

xcompile: test
	@rm -rf build/
	@mkdir -p build/
	gox \
		-os="darwin" \
		-os="linux" \
		-os="freebsd" \
		-os="openbsd" \
		-os="solaris" \
		-output="build/$(NAME)_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)"

package: xcompile
	@mkdir -p build/tgz
	for f in $(shell find build -name converge | cut -d/ -f2); do \
	  (cd $(shell pwd)/build/$$f && tar -zcvf ../tgz/$$f.tar.gz converge); \
    echo $$f; \
  done

.PHONY: test xcompile package
