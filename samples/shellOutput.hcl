/* This task demonstrates how shell output is handled when running
'tasks', specificially the ways that stderr and stdout are written to
the screen.  Running this example with 'test.txt' present will result
in 'check found a test file' being found, however running without this
file will result only in 'check found no test file' being printed-
this illustrates how secondary runs of the check script are supressed
during execution.  */

task "shell output" {
  interpreter = "/bin/bash"
  check_flags = ["-n"]

  check = <<END
if [[ -f test.txt ]]; then
echo -n "check found a test file"
else
echo -n "check found no test file"
fi

[[ -f test.txt ]]
END

  apply = <<END
echo -n "to stdout"
echo -n "to stderr" >&2
touch test.txt
END
}
