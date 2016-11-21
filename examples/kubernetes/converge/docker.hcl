param "docker-package" {
  default = "docker-engine"
}

param "docker-service" {
  default = "docker"
}

param "docker-repo" {
  default = "deb https://apt.dockerproject.org/repo ubuntu-xenial main"
}

param "docker-repo-keyid" {
  default = "F76221572C52609D"
}

param "docker-repo-recv-key" {
  default = "58118E89F3A912897C070ADBF76221572C52609D"
}

param "docker-repo-keyserver" {
  default = "hkp://p80.pool.sks-keyservers.net:80"
}

task "docker-repo-key" {
  check = "apt-key adv --list-public-keys --keyid-format=long | grep {{param `docker-repo-keyid`}}"
  apply = "apt-key adv --keyserver {{param `docker-repo-keyserver`}} --recv-keys {{param `docker-repo-recv-key`}}"
}

file.content "docker-repo" {
  destination = "/etc/apt/sources.list.d/docker.list"
  content     = "{{param `docker-repo`}}"
}

file.mode "docker-repo" {
  destination = "{{lookup `file.content.docker-repo.destination`}}"
  mode        = "0755"
}

task.query "apt-get-update" {
  query   = "apt-get update -y || true"
  depends = ["task.docker-repo-key", "file.content.docker-repo"]
}

package.apt "docker-install" {
  name    = "{{param `docker-package`}}"
  state   = "present"
  depends = ["task.query.apt-get-update"]
  group   = "apt"
}

task "docker-enable" {
  check   = "systemctl is-enabled {{param `docker-service`}}"
  apply   = "systemctl daemon-reload; systemctl enable {{param `docker-service`}}"
  depends = ["package.apt.docker-install"]
}

task "docker-start" {
  check   = "systemctl is-active {{param `docker-service`}}"
  apply   = "systemctl daemon-reload; systemctl start {{param `docker-service`}}"
  depends = ["task.docker-enable"]
}
