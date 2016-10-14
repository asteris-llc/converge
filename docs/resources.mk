# automatically generated Wed Oct 12 11:29:51 CDT 2016
include Makefile

content/resources/docker.container.md: extract ../resource/docker/container/preparer.go ../samples/dockerContainer.hcl
	echo '---' > $@
	echo 'title: "docker.container"' >> $@
	echo 'slug: "docker-container"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/dockerContainer.hcl --resource-name docker.container --path ../resource/docker/container/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/docker.image.md: extract ../resource/docker/image/preparer.go ../samples/dockerImage.hcl
	echo '---' > $@
	echo 'title: "docker.image"' >> $@
	echo 'slug: "docker-image"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/dockerImage.hcl --resource-name docker.image --path ../resource/docker/image/preparer.go --type Preparer --strip-doc-lines=2 >> $@

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

content/resources/file.directory.md: extract ../resource/file/directory/preparer.go ../samples/fileDirectory.hcl
	echo '---' > $@
	echo 'title: "file.directory"' >> $@
	echo 'slug: "file-directory"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/fileDirectory.hcl --resource-name file.directory --path ../resource/file/directory/preparer.go --type Preparer --strip-doc-lines=2 >> $@

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

content/resources/task.query.md: extract ../resource/shell/query/preparer.go ../samples/query.hcl
	echo '---' > $@
	echo 'title: "task.query"' >> $@
	echo 'slug: "task-query"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/query.hcl --resource-name task.query --path ../resource/shell/query/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/user.group.md: extract ../resource/group/preparer.go ../samples/group.hcl
	echo '---' > $@
	echo 'title: "user.group"' >> $@
	echo 'slug: "user-group"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/group.hcl --resource-name user.group --path ../resource/group/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/user.user.md: extract ../resource/user/preparer.go ../samples/user.hcl
	echo '---' > $@
	echo 'title: "user.user"' >> $@
	echo 'slug: "user-user"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/user.hcl --resource-name user.user --path ../resource/user/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/wait.query.md: extract ../resource/wait/preparer.go ../samples/wait.hcl
	echo '---' > $@
	echo 'title: "wait.query"' >> $@
	echo 'slug: "wait-query"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/wait.hcl --resource-name wait.query --path ../resource/wait/preparer.go --type Preparer --strip-doc-lines=2 >> $@

content/resources/wait.port.md: extract ../resource/wait/port/preparer.go ../samples/waitPort.hcl
	echo '---' > $@
	echo 'title: "wait.port"' >> $@
	echo 'slug: "wait-port"' >> $@
	echo "date: \"$$(date -j '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\1/')\"" >> $@
	echo "menu:" >> $@
	echo "  main:" >> $@
	echo "    parent: resources" >> $@
	echo '---' >> $@
	echo >> $@
	./extract --example ../samples/waitPort.hcl --resource-name wait.port --path ../resource/wait/port/preparer.go --type Preparer --strip-doc-lines=2 >> $@

