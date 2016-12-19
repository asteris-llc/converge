# the funky $(eval ...) here and below is a sneaky Make trick. The = allows for
# recursive (lazy) expansion, but the first time it's expanded we eval to
# replace the value with a expand-once (strict) variable. Variables defined in
# this way will be evaluated exactly once, and only if used.
#
# meta information about the project
REPO = $(eval REPO := $(shell go list -f '{{.ImportPath}}' .))$(value REPO)
NAME = $(eval NAME := $(shell basename ${REPO}))$(value NAME)
VERSION = $(eval VERSION := $(shell git describe --dirty))$(value VERSION)
PACKAGE_VERSION = $(eval PACKAGE_VERSION := $(shell echo ${VERSION} | sed 's/-dirty//g'))$(value PACKAGE_VERSION)

# sources to evaluate
SRCDIRS = $(eval SRCDIRS := $(shell glide novendor --no-subdir | grep -v '^.$$' | sed 's|/$$||g'))$(value SRCDIRS)
SRCFILES = $(eval SRCFILES := main.go $(shell find ${SRCDIRS} -name '*.go'))$(value SRCFILES)

# binaries
converge: vendor ${SRCFILES} rpc/pb/root.pb.go rpc/pb/root.pb.gw.go
	go build -ldflags="-X ${REPO}/cmd.Version=${PACKAGE}"

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

# vendoring
vendor: glide.yaml glide.lock
	glide install
	make vendor-clean

vendor-clean:
	find vendor -not -name '*.go' -not -name '*.s' -not -name '*.pl' -not -name '*.c' -not -name LICENSE -not -name '*.proto' -type f -delete

# testing
test: gotest validate-samples validate-error-samples blackbox

gotest:
	go test $(shell glide novendor)

validate-samples: converge samples/*.hcl
	@echo
	@echo === checking validity of all samples ===
	./converge validate samples/*.hcl
	@echo
	@echo === checking formatting of all samples ===
	./converge fmt --check samples/*.hcl

validate-error-samples: samples/errors/*.hcl

samples/errors/*.hcl: converge
	@echo
	@echo === validating $@ should fail ===
	./converge validate $@ || exit 0 && exit 1

blackbox: blackbox/*.sh

blackbox/*.sh: converge
	@echo
	@echo === testing $@ ===
	@$@

# fuzzing
fuzzing/*:
	@echo
	@echo === fuzzing $(shell basename $@) ===
	@cd $@ && ./run.sh
	@test -d $@/crashers && test "$$(find $@/crashers -type f)" = "" || (echo found crashers; tail -n 100000 $@/crashers/*; exit 1)

# benchmarks
BENCH := .
BENCHDIRS = $(shell find ${SRCDIRS} -name '*_test.go' | xargs grep '*testing.B' | cut -d: -f1 | xargs -n1 dirname | uniq)
bench:
	go test -run '^$$' -bench=${BENCH} -benchmem ${BENCHDIRS}

# linting
LINTDIRS = $(eval LINTDIRS := $(shell find ${SRCDIRS} -type d -not -path './rpc/pb' -not -path './docs*'))$(value LINTDIRS)
lint:
	@echo '=== golint ==='
	@for dir in ${LINTDIRS}; do golint $${dir}; done # github.com/golang/lint/golint

	@echo '=== gosimple ==='
	@gosimple ${LINTDIRS} # honnef.co/go/simple/cmd/gosimple

	@echo '=== unconvert ==='
	@unconvert ${LINTDIRS} # github.com/mdempsky/unconvert

	@echo '=== structcheck ==='
	@structcheck ${LINTDIRS} # github.com/opennota/check/cmd/structcheck

	@echo '=== varcheck ==='
	@varcheck ${LINTDIRS} # github.com/opennota/check/cmd/varcheck

	@echo '=== aligncheck ==='
	@aligncheck ${LINTDIRS} # github.com/opennota/check/cmd/aligncheck

	@echo '=== gas ==='
	@gas ${LINTDIRS} # github.com/HewlettPackard/gas

# documentation
samples/%.png: samples/% converge
	@echo
	@echo === rendering $@ ===
	./converge graph --local $< | dot -Tpng -o$@

docs/public:
	cd docs; make public

# packaging
xcompile: rpc/pb/root.pb.go rpc/pb/root.pb.gw.go test
	@echo "set version to ${PACKAGE_VERSION}"

	@rm -rf build/
	@mkdir -p build/
	gox \
			-ldflags="-X ${REPO}/cmd.Version=${PACKAGE_VERSION} -s -w" \
			-osarch="darwin/386" \
			-osarch="darwin/amd64" \
			-os="linux" \
			-os="freebsd" \
			-os="solaris" \
			-output="build/${NAME}_${PACKAGE_VERSION}_{{.OS}}_{{.Arch}}/${NAME}"
	find build -type f -execdir /bin/bash -c 'shasum -a 256 $$0 > $$0.sha256sum' \{\} \;

package: xcompile
	@mkdir -p build/tgz
	for f in $(shell find build -name converge | cut -d/ -f2); do \
			(cd $(shell pwd)/build/$$f && tar -zcvf ../tgz/$$f.tar.gz *); \
		done
	(cd build/tgz; shasum -a 512 * > tgz.sha512sum)

.PHONY: test gotest validate-samples validate-error-samples samples/errors/*.hcl blackbox/*.sh fuzzing/* bench lint xcompile package
