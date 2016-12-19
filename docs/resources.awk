BEGIN {
    FS=",";
    root="content/resources"

    print "RESOURCES ="
    print root ":";
    print "\tmkdir -p " root;
    print "";
}
{
    name = $1;
    source = $2;
    example = $3;
    type = $4;
    tasksource = $5;
    task = $6;
    resource=root "/" name ".md"

    slug = name;
    gsub("\\.", "-", slug);

    print "RESOURCES += " resource
    print resource ": sources.csv extract " root " " source " " example " " tasksource;
    print "\techo '---' > $@"
    print "\techo 'title: \"" name "\"' >> $@"
    print "\techo 'slug: \"" slug "\"' >> $@"
    print "\techo \"date: \\\"$$(date '+%Y-%m-%dT%H:%M:%S%z' | sed -E 's/(..)$$/:\\1/')\\\"\" >> $@"
    print "\techo \"menu:\" >> $@"
    print "\techo \"  main:\" >> $@"
    print "\techo \"    parent: resources\" >> $@"
    print "\techo '---' >> $@"
    print "\techo >> $@"
    print "\t./extract --example " example " --resource-name " name " --path " source " --type " type " --task " task " --task-path " tasksource " --strip-doc-lines=2 >> $@"
    print "";
}
