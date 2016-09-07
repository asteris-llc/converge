NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' cmd/root.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' cmd/version.go)
RPCLINT=$(shell find ./rpc -type f \( -not -iname 'root.*.go' -iname '*.go' \) )
TOLINT = $(shell find . -type f \( -not -ipath './vendor*'  -not -ipath './docs_source*' -not -ipath './rpc*' -not -iname 'main.go' -iname '*.go' \) -exec dirname {} \; | sort -u)
TESTDIRS = $(shell find . -name '*_test.go' -exec dirname \{\} \; | grep -v vendor | uniq)
NONVENDOR = ${shell find . -name '*.go' | grep -v vendor}
BENCHDIRS= $(shell find . -name '*_test.go' | grep -v vendor | xargs grep '*testing.B' | cut -d: -f1 | xargs dirname | uniq)
BENCH = .

converge: $(shell find . -name '*.go') rpc/pb/root.pb.go rpc/pb/root.pb.gw.go
	go build -ldflags="-s -w" .

test: converge gotest samples/*.hcl samples/errors/*.hcl blackbox/*.sh
	@echo
	@echo === check validity of all samples ===
	./converge validate samples/*.hcl
	@echo
	@echo === check formatting of all samples ===
	./converge fmt --check samples/*.hcl

gotest:
	go test ${TESTDIRS}

license-check:
	@echo "=== Missing License Files ==="
	@./check_license.sh

bench:
	go test -run='^$$' -bench=${BENCH} -benchmem ${BENCHDIRS}

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
	./converge graph --local $< | dot -Tpng -o$@

lint: rpclint
	@echo '# golint'
	@for dir in ${TOLINT}; do golint $${dir}/...; done # github.com/golang/lint/golint
	@for file in ${RPCLINT}; do golint $${file}; done # github.com/golang/lint/golint

	@echo '# go tool vet'
	@go tool vet -all -shadow ${TOLINT}
	@go tool vet -all -shadow ${RPCLINT} # built in

	@echo '# gosimple'
	@gosimple ${TOLINT} # honnef.co/go/simple/cmd/gosimple
	@gosimple ${RPCLINT} # honnef.co/go/simple/cmd/gosimple

	@echo '# unconvert'
	@unconvert ${TOLINT} # github.com/mdempsky/unconvert
	@unconvert ${RPCLINT} # github.com/mdempsky/unconvert

	@echo '# structcheck'
	@structcheck ${TOLINT} # github.com/opennota/check/cmd/structcheck
	@structcheck ${RPCLINT} # github.com/opennota/check/cmd/structcheck

	@echo '# varcheck'
	@varcheck ${TOLINT} # github.com/opennota/check/cmd/varcheck
	@varcheck ${RPCLINT} # github.com/opennota/check/cmd/varcheck

	@echo '# aligncheck'
	@aligncheck ${TOLINT} # github.com/opennota/check/cmd/aligncheck
	@aligncheck ${RPCLINT} # github.com/opennota/check/cmd/aligncheck

	@echo '# gas'
	@gas ${TOLINT} # github.com/HewlettPackard/gas
	@gas ${RPCLINT} # github.com/HewlettPackard/gas

vendor: ${NONVENDOR}
	glide install --strip-vcs --strip-vendor --update-vendored
	make vendor-clean

vendor-update: ${NOVENDOR}
	glide update --strip-vcs --strip-vendor --update-vendored
	make vendor-clean

vendor-clean: ${NOVENDOR}
	find vendor -not -name '*.go' -not -name '*.s' -not -name '*.pl' -not -name '*.c' -not -name LICENSE -not -name '*.proto' -type f -delete

xcompile: rpc/pb/root.pb.go rpc/pb/root.pb.gw.go test
	@rm -rf build/
	@mkdir -p build/
	gox \
    -ldflags="-s -w" \
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

rpc/pb/root.pb.go: rpc/pb/root.proto
	protoc -I rpc/pb \
	 -I vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	 --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:rpc/pb \
	 rpc/pb/root.proto

rpc/pb/root.pb.gw.go: rpc/pb/root.proto
	protoc -I rpc/pb \
	 -I vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	 --grpc-gateway_out=logtostderr=true:rpc/pb \
	 rpc/pb/root.proto

rpc/pb/root.swagger.json: rpc/pb/root.proto
	protoc -I rpc/pb \
	 -I vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	 --swagger_out=logtostderr=true:rpc/pb \
	 rpc/pb/root.proto

docs: docs_source/**/*
	rm -rf docs || true
	$(MAKE) -C docs_source
	mv docs_source/public docs

.PHONY: test gotest vendor-update vendor-clean xcompile package samples/errors/*.hcl blackbox/*.sh lint bench license-check
