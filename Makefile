.PHONY = test

converge: main.go cmd/* load/* resource/* exec/* vendor/**/*
	go build .

test: converge samples/*.hcl
	go test -v ./...
	find samples -type f -name '*.hcl' -exec ./converge check \{\} \;

samples/%.png: samples/% converge
	./converge graph $< | dot -Tpng -o$@

vendor: main.go cmd/* load/* resource/* exec/*
	godep save -t ./...
