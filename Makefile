.PHONY = test
TESTDIRS = $(shell find . -name '*_test.go' | cut -d/ -f1-2 | uniq | grep -v vendor)

converge: main.go cmd/* load/* resource/* vendor/**/*
	go build .

test: converge samples/*.hcl
	go test -v ${TESTDIRS}
	find samples -type f -name '*.hcl' -exec ./converge check \{\} \;

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@

vendor: main.go cmd/* load/* resource/* exec/*
	godep save -t ./...
