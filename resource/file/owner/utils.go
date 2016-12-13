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

package owner

import "fmt"

func normalizeUser(p OSProxy, username, uid string) (string, string, error) {
	return normalizeTuple(p, username, uid, usernameFromUID, uidFromUsername)
}

func normalizeGroup(p OSProxy, groupname, gid string) (string, string, error) {
	return normalizeTuple(p, groupname, gid, groupnameFromGID, gidFromGroupname)
}

func normalizeTuple(p OSProxy, fst, snd string, deriveFirst, deriveSnd func(OSProxy, string) (string, error)) (string, string, error) {
	var err error
	if fst == "" {
		if snd == "" {
			return "", "", nil
		}
		fst, err = deriveFirst(p, snd)
	} else {
		snd, err = deriveSnd(p, fst)
	}
	return fst, snd, err
}

func usernameFromUID(p OSProxy, uid string) (string, error) {
	u, err := p.LookupId(uid)
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

func uidFromUsername(p OSProxy, name string) (string, error) {
	u, err := p.Lookup(name)
	if err != nil {
		return "", err
	}
	return u.Uid, nil
}

func groupnameFromGID(p OSProxy, gid string) (string, error) {
	u, err := p.LookupGroupId(gid)
	if err != nil {
		return "", err
	}
	return u.Name, nil
}

func gidFromGroupname(p OSProxy, name string) (string, error) {
	u, err := p.LookupGroup(name)
	if err != nil {
		return "", err
	}
	return u.Gid, nil
}

func show(i interface{}) string {
	return fmt.Sprintf("%v", i)
}
