.PHONY = test _testgo _testcheck

converge: main.go cmd/* load/* resource/*
	go build .

test: converge samples/*
	go test -v ./...
	find samples -type f -exec ./converge check \{\} \;
