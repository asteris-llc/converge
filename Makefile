NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' cmd/root.go)
RPCLINT=$(shell find ./rpc -type f \( -not -iname 'root.*.go' -iname '*.go' \) )
TOLINT = $(shell find . -type f \( -not -ipath './vendor*'  -not -ipath './docs*' -not -ipath './rpc*' -not -iname 'main.go' -iname '*.go' \) -exec dirname {} \; | sort -u)
TESTDIRS = $(shell find . -name '*_test.go' -exec dirname \{\} \; | grep -v vendor | uniq)
NONVENDOR = ${shell find . -name '*.go' | grep -v vendor}
BENCHDIRS= $(shell find . -name '*_test.go' | grep -v vendor | xargs grep '*testing.B' | cut -d: -f1 | xargs -n1 dirname | uniq)
BENCH = .
REPO = github.com/asteris-llc/converge

converge: $(shell find . -name '*.go') rpc/pb/root.pb.go rpc/pb/root.pb.gw.go
	go build -ldflags="-X ${REPO}/cmd.Version=$(shell git describe --dirty) -s -w" .

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

lint:
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
	@echo "set version to $(shell git describe)"

	@rm -rf build/
	@mkdir -p build/
	gox \
    -ldflags="-X ${REPO}/cmd.Version=$(shell git describe) -s -w" \
		-osarch="darwin/386" \
		-osarch="darwin/amd64" \
		-os="linux" \
		-os="freebsd" \
		-os="solaris" \
		-output="build/$(NAME)_$(shell git describe)_{{.OS}}_{{.Arch}}/$(NAME)"
	find build -type file -execdir /bin/bash -c 'shasum -a 256 $$0 > $$0.sha256sum' \{\} \;

package: xcompile
	@mkdir -p build/tgz
	for f in $(shell find build -name converge | cut -d/ -f2); do \
		(cd $(shell pwd)/build/$$f && tar -zcvf ../tgz/$$f.tar.gz *); \
	done
	(cd build/tgz; shasum -a 512 * > tgz.sha256sum)

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

.PHONY: test gotest vendor-update vendor-clean xcompile package samples/errors/*.hcl blackbox/*.sh lint bench license-check
