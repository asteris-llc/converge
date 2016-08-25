# automatically generated Wed Aug 24 23:31:55 CDT 2016
include Makefile

content/resources/file.content.md: extract ../resource/file/content/preparer.go ../samples/fileContent.hcl
	echo '---' > $@
	echo 'title: "file.content"' >> $@
	echo 'slug: "file-content"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/fileContent.hcl --resource-name file.content --path ../resource/file/content/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/file.mode.md: extract ../resource/file/mode/preparer.go ../samples/fileMode.hcl
	echo '---' > $@
	echo 'title: "file.mode"' >> $@
	echo 'slug: "file-mode"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/fileMode.hcl --resource-name file.mode --path ../resource/file/mode/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/module.md: extract ../resource/module/preparer.go ../samples/sourceFile.hcl
	echo '---' > $@
	echo 'title: "module"' >> $@
	echo 'slug: "module"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/sourceFile.hcl --resource-name module --path ../resource/module/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/param.md: extract ../resource/param/preparer.go ../samples/basic.hcl
	echo '---' > $@
	echo 'title: "param"' >> $@
	echo 'slug: "param"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/basic.hcl --resource-name param --path ../resource/param/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/task.md: extract ../resource/shell/preparer.go ../samples/basic.hcl
	echo '---' > $@
	echo 'title: "task"' >> $@
	echo 'slug: "task"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/basic.hcl --resource-name task --path ../resource/shell/preparer.go --type Preparer --strip-doc-lines=2 >> $@

