.PHONY = test _testgo _testcheck

converge: main.go cmd/* load/* resource/* exec/*
	go build .

test: converge samples/*
	go test -v ./...
	find samples -type f -name '*.hcl' -exec ./converge check \{\} \;

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@
