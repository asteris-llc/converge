all: public

-include resources.mk

public: content/commands content/*.md static/**/* ${RESOURCES} content/license.md static/images/**/*.png
	hugo

publish: public
	./publish.sh $$VERSION

# automatic documentation extraction
extract: extract.go
	go build -o $@ extract.go

resources.mk: resources.awk sources.csv
	awk -f resources.awk sources.csv > $@

content/commands: ../cmd/*.go
	go run ../main.go gen docs --path=$@

# content to copy
content/license.md: ../LICENSE
	@mkdir content/license || true
	echo '---' > $@
	echo 'title: "License"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    weight: 10" >> $@
	echo '---' >> $@
	echo >> $@
	cat ../LICENSE >> $@

# images
converge:
	go build ..

static/images/dependencies/with-depends.png: static/images/dependencies/with-depends.hcl converge
	./converge graph --local $< | dot -Tpng -o$@

static/images/dependencies/without-depends.png: static/images/dependencies/without-depends.hcl converge
	./converge graph --local $< | dot -Tpng -o$@

static/images/getting-started/hello-world.png: static/images/getting-started/hello-world.hcl converge
	./converge graph --local --show-params $< | dot -Tpng -o$@

static/images/getting-started/hello-world-params.png: static/images/getting-started/hello-world-params.hcl converge
	./converge graph --local --show-params $< | dot -Tpng -o$@

static/images/getting-started/hello-you.png: static/images/getting-started/hello-you.hcl converge
	./converge graph --local --show-params $< | dot -Tpng -o$@
