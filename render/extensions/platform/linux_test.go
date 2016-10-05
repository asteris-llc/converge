// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package platform

import (
	"reflect"
	"testing"
)

func TestParseLSBAlpine(t *testing.T) {
	content := `NAME="Alpine Linux"
ID=alpine
VERSION_ID=3.4.0
PRETTY_NAME="Alpine Linux v3.4"
HOME_URL="http://alpinelinux.org"
BUG_REPORT_URL="http://bugs.alpinelinux.org"
`

	var platform Platform
	var expected Platform

	expected.Name = "Alpine Linux"
	expected.LinuxDistribution = "alpine"
	expected.Version = "3.4.0"
	expected.Build = ""

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)
}

func TestParseLSBCentOS(t *testing.T) {
	content := `NAME="CentOS Linux"
VERSION="7 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:7"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-7"
CENTOS_MANTISBT_PROJECT_VERSION="7"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="7"

`
	var platform Platform
	var expected Platform

	expected.Name = "CentOS Linux"
	expected.LinuxDistribution = "centos"
	expected.Version = "7"
	expected.LinuxLSBLike = []string{"rhel", "fedora"}

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)
}

func TestParseLSBCoreOS(t *testing.T) {
	content := `NAME=CoreOS
ID=coreos
VERSION=835.9.0
VERSION_ID=835.9.0
BUILD_ID=
PRETTY_NAME="CoreOS 835.9.0"
ANSI_COLOR="1;32"
HOME_URL="https://coreos.com/"
BUG_REPORT_URL="https://github.com/coreos/bugs/issues"`
	var platform Platform
	var expected Platform

	expected.Name = "CoreOS"
	expected.LinuxDistribution = "coreos"
	expected.Version = "835.9.0"

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)
}

func TestParseLSBDebian(t *testing.T) {
	content := `PRETTY_NAME="Debian GNU/Linux 8 (jessie)"
NAME="Debian GNU/Linux"
VERSION_ID="8"
VERSION="8 (jessie)"
ID=debian
HOME_URL="http://www.debian.org/"
SUPPORT_URL="http://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"
`

	var platform Platform
	var expected Platform

	expected.Name = "Debian GNU/Linux"
	expected.LinuxDistribution = "debian"
	expected.Version = "8"

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)

}

func TestParseLSBNixOS(t *testing.T) {
	content := `NAME=NixOS
ID=nixos
VERSION="16.09.git.bfc0c28 (Flounder)"
VERSION_ID="16.09.git.bfc0c28"
PRETTY_NAME="NixOS 16.09.git.bfc0c28 (Flounder)"
HOME_URL="http://nixos.org/"
`
	var platform Platform
	var expected Platform

	expected.Name = "NixOS"
	expected.LinuxDistribution = "nixos"
	expected.Version = "16.09.git.bfc0c28"

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)

}

func TestParseLSBUbuntu(t *testing.T) {
	content := `NAME="Ubuntu"
VERSION="16.04.1 LTS (Xenial Xerus)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 16.04.1 LTS"
VERSION_ID="16.04"
HOME_URL="http://www.ubuntu.com/"
SUPPORT_URL="http://help.ubuntu.com/"
BUG_REPORT_URL="http://bugs.launchpad.net/ubuntu/"
UBUNTU_CODENAME=xenial
`

	var platform Platform
	var expected Platform

	expected.Name = "Ubuntu"
	expected.LinuxDistribution = "ubuntu"
	expected.Version = "16.04"
	expected.LinuxLSBLike = []string{"debian"}

	platform.ParseLSBContent(content)
	ComparePlatform(t, expected, platform)

}

func ComparePlatform(t *testing.T, expected Platform, platform Platform) {

	if platform.Name != expected.Name {
		t.Errorf("ParseLSBContent Name: wanted %q Linux, got %q\n", expected.Name, platform.Name)
	}

	if platform.LinuxDistribution != expected.LinuxDistribution {
		t.Errorf("ParseLSBContent Name: wanted %q Linux, got %q\n", expected.LinuxDistribution, platform.LinuxDistribution)
	}

	if platform.Version != expected.Version {
		t.Errorf("ParseLSBContent Version: wanted %q, got %q\n", expected.Version, platform.Version)
	}

	if platform.Build != expected.Build {
		t.Errorf("ParseLSBContent Build: wanted %q, got %q\n", expected.Build, platform.Build)
	}

	if !reflect.DeepEqual(expected.LinuxLSBLike, platform.LinuxLSBLike) {
		t.Errorf("ParseLSBContent LinuxLSBLike wanted %q got %q\n", expected.LinuxLSBLike, platform.LinuxLSBLike)
	}

}
