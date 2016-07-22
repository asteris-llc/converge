NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' cmd/root.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' cmd/version.go)
TESTDIRS = $(shell find . -name '*_test.go' -exec dirname \{\} \; | grep -v vendor | uniq)
PNGS = $(shell find samples -name '*.hcl' | grep -v errors | awk '{ print $$1".png" }')
NONVENDOR = ${shell find . -name '*.go' | grep -v vendor}

converge: $(shell find . -name '*.go')
	go build .

test: converge gotest samples/*.hcl samples/errors/*.hcl blackbox/*.sh ${PNGS}
	@echo
	@echo === check validity of all samples ===
	./converge validate samples/*.hcl
	@echo
	@echo === check formatting of all samples ===
	./converge fmt --check samples/*.hcl

gotest:
	go test -v ${TESTDIRS}

samples/errors/*.hcl: converge
	@echo
	@echo === validating $@ should fail ==
	./converge validate $@ || exit 0 && exit 1

blackbox/*.sh: converge
	@echo
	@echo === testing $@ ===
	@$@

samples/%.png: samples/% converge
	@echo
	@echo === rendering $@ ===
	./converge graph $< | dot -Tpng -o$@

vendor: ${NONVENDOR}
	glide install --strip-vcs --strip-vendor --update-vendored
	find vendor -not -name '*.go' -not -name '*.s' -not -name '*.pl' -not -name '*.c' -not -name LICENSE -type f -delete

vendor-update: ${NOVENDOR}
	glide update --strip-vcs --strip-vendor --update-vendored
	find vendor -not -name '*.go' -not -name LICENSE -type f -delete

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

.PHONY: test gotest vendor-update xcompile package samples/errors/*.hcl blackbox/*.sh
