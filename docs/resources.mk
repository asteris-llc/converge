# automatically generated Wed Aug 24 16:45:02 CDT 2016
include Makefile

content/resources/file.content.md: extract ../resource/file/content/content.go ../samples/fileContent.hcl
	echo '---' > $@
	echo 'title: "file.content"' >> $@
	echo 'slug: "file-content"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/fileContent.hcl --resource-name file.content --path ../resource/file/content/content.go --type Content >> $@

content/resources/file.mode.md: extract ../resource/file/mode/mode.go ../samples/fileMode.hcl
	echo '---' > $@
	echo 'title: "file.mode"' >> $@
	echo 'slug: "file-mode"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/fileMode.hcl --resource-name file.mode --path ../resource/file/mode/mode.go --type Mode >> $@

content/resources/module.md: extract ../resource/module/module.go ../samples/basic.hcl
	echo '---' > $@
	echo 'title: "module"' >> $@
	echo 'slug: "module"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/basic.hcl --resource-name module --path ../resource/module/module.go --type Module >> $@

content/resources/param.md: extract ../resource/param/param.go ../samples/basic.hcl
	echo '---' > $@
	echo 'title: "param"' >> $@
	echo 'slug: "param"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/basic.hcl --resource-name param --path ../resource/param/param.go --type Param >> $@

content/resources/task.md: extract ../resource/shell/shell.go ../samples/basic.hcl
	echo '---' > $@
	echo 'title: "task"' >> $@
	echo 'slug: "task"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/basic.hcl --resource-name task --path ../resource/shell/shell.go --type Shell >> $@

