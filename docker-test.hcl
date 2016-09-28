param "test-case" {}
param "converge-bin-dir" {}
param "rpc-token" {}

param "image-name" {
	default = "ubuntu"
}
param "image-tag" {
	default = "latest"
	}

docker.image "converge" {
	name = "{{param `image-name`}}"
	tag = "{{param `image-tag`}}"
}

docker.container "converge" {
	name = "converge-test"
	image = "{{param `image-name`}}:{{param `image-tag`}}"

	expose = ["4774"]
	volumes = [
		"{{param `converge-bin-dir`}}:/converge",
		"{{env `PWD`}}/samples:/samples",
		"{{env `PWD`}}/examples:/examples"
	]
	entrypoint = "/converge/converge"
	command = ["server", "--rpc-token", "{{param `rpc-token`}}"]

	depends = ["docker.image.converge"]
}

task.query "wait" {
	query = "sleep 10"
	depends = ["task.container.converge"]
}

task.query "container-port" {
	query = "docker inspect {{lookup `docker.container.converge.name`}} | jq '.NetworkSettings.Ports[\"4775/tcp\"][0].HostPort'"
	depends = ["task.query.wait"]
}

task "run-test" {
	check = "docker exec {{lookup `docker.container.converge.name`}} test -f ~/{{param `test-case`}}-completed"
	apply = <<EOF
{{$name := lookup `docker.container.converge.name`}}
docker exec {{$name}} apply --rpc-addr :4774 --rpc-token {{param `rpc-token`}} {{param `test-case`}}
./converge apply --rpc-token {{param `rpc-token`}} --rpc-addr {{$name}} {{param `test-case`}}
docker exec {{$name}} touch ~/{{param `test-case`}}-completed
EOF
	depends = ["task.query.container-port"]
}

task "destroy-container" {
	check = <<EOF
docker ps -a | grep -q {{lookup `docker.container.converge.name`}}
if [ $? -ne 0 ]; then exit 1; else exit 0; fi
EOF
	apply = "docker rm -f {{lookup `docker.container.converge.name`}}"

	depends = ["task.run-test"]
}

task "destroy-image" {
	check = <<EOF
docker images | grep -q {{lookup `docker.container.converge.image`}}
if [ $? -ne 0 ]; then exit 1; else exit 0; fi
EOF
	apply = "docker rmi image"

	depends = ["task.destroy-container"]
}
