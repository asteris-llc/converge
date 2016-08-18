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
