.PHONY = test
TESTDIRS = $(shell find . -name '*_test.go' | cut -d/ -f1-2 | uniq | grep -v vendor)

converge: main.go cmd/* load/* resource/* vendor/**/*
	go build .

test: converge samples/*.hcl samples/errors/*.hcl
	go test -v ${TESTDIRS}
	find samples -depth 1 -type f -name '*.hcl' | xargs ./converge validate
	find samples -depth 1 -type f -name '*.hcl' | xargs ./converge fmt --check

samples/errors/*.hcl: converge
	@echo === planning $@ should fail ===
	./converge plan $@ || exit 0 && exit 1

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@

vendor: main.go cmd/* load/* resource/* exec/*
	godep save -t ./...
