.PHONY = test
TESTDIRS = $(shell find . -name '*_test.go' | cut -d/ -f1-2 | uniq | grep -v vendor)

converge: main.go cmd/* load/* resource/* vendor/**/*
	go build .

test: converge samples/*.hcl samples/errors/*.hcl
	go test -v ${TESTDIRS}
	./converge validate samples/*.hcl
	./converge fmt --check samples/*.hcl

samples/errors/*.hcl: converge
	@echo === planning $@ should fail ===
	./converge plan $@ || exit 0 && exit 1

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@

vendor: main.go cmd/* load/* resource/* exec/*
	godep save -t ./...
