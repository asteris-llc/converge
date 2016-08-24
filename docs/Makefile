TASKFILE=../resource/resource.go
TASKLOC=${TASKFILE}:\#$(shell grep -b 'type Task interface' ${TASKFILE} | cut -d: -f1)

public: content/modules
	hugo

extract: extract.go
	go build -o extract extract.go

content/modules: modules.mk $(shell cut -d, -f2,3 sources.csv | sed -E 's|,| |g')
	mkdir -p content/modules || true
	@for line in $$(cat sources.csv); do \
		make -f modules.mk content/modules/$$(echo $$line | cut -d, -f1).md; \
	done

modules.mk: sources.csv moduleContent.mk.tmpl
	@echo "# automatically generated $$(date)" > modules.mk
	@echo "include Makefile" >> modules.mk
	@echo >> modules.mk
	@for line in $$(cat sources.csv); do \
    sed -e "s|@@NAME|$$(echo $$line | cut -d, -f1 | tr -d '\n')|g" \
        -e "s|@@SOURCE|$$(echo $$line | cut -d, -f2 | tr -d '\n')|g" \
        -e "s|@@EXAMPLE|$$(echo $$line | cut -d, -f3 | tr -d '\n')|g" \
        -e "s|@@TYPE|$$(echo $$line | cut -d, -f4 | tr -d '\n')|g" \
        -e "s|@@TARGET|content/modules/$$(echo $$line | cut -d, -f1 | tr -d '\n').md|g" \
        moduleContent.mk.tmpl >> modules.mk; \
    echo >> modules.mk; \
  done
