TASKFILE=../resource/resource.go
TASKLOC=${TASKFILE}:\#$(shell grep -b 'type Task interface' ${TASKFILE} | cut -d: -f1)

public: content/*.md static/**/* content/resources content/license.md static/images/**/*.png
	hugo

publish:
	./publish.sh $$VERSION

extract: extract.go
	go build -o $@ extract.go

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

content/resources: extract resources.mk $(shell cut -d, -f2,3 sources.csv | sed -E 's|,| |g')
	mkdir -p $@ || true
	@for line in $$(cat sources.csv); do \
		make -f resources.mk $@/$$(echo $$line | cut -d, -f1).md; \
	done

resources.mk: sources.csv resourceContent.mk.tmpl
	@echo "# automatically generated $$(date)" > $@
	@echo "include Makefile" >> $@
	@echo >> $@
	@for line in $$(cat sources.csv); do \
    sed -e "s|@@NAME|$$(echo $$line | cut -d, -f1 | tr -d '\n')|g" \
        -e "s|@@SLUG|$$(echo $$line | cut -d, -f1 | tr -d '\n' | sed 's/\./-/g')|g" \
        -e "s|@@SOURCE|$$(echo $$line | cut -d, -f2 | tr -d '\n')|g" \
        -e "s|@@EXAMPLE|$$(echo $$line | cut -d, -f3 | tr -d '\n')|g" \
        -e "s|@@TYPE|$$(echo $$line | cut -d, -f4 | tr -d '\n')|g" \
        -e "s|@@TARGET|content/resources/$$(echo $$line | cut -d, -f1 | tr -d '\n').md|g" \
        resourceContent.mk.tmpl >> $@; \
    echo >> $@; \
	done

# Images

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

.PHONY=publish
