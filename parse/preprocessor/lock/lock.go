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

package lock

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/parse"
)

// we might want to change the keywords later, so keep them in a map, later we
// can replace map lookups with the final keyword
var keywords = map[string]string{
	"lock":   "lock.lock",
	"unlock": "lock.unlock",
}

// Lock encapsulates lock entry and exit nodes
type Lock struct {
	LockID     string
	UnlockID   string
	LockNode   *parse.Node
	UnlockNode *parse.Node
}

// NewLock generates lock nodes
func NewLock(n *parse.Node) (*Lock, error) {
	name, err := getLock(n)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, nil
	}

	genNode := func(name, lockType string) (*parse.Node, error) {
		lockNode, generr := parse.Parse([]byte(fmt.Sprintf("%s \"%s\" {}", keywords[lockType], name)))
		if generr != nil {
			return nil, generr
		}
		return lockNode[0], nil
	}

	lockNode, err := genNode(name, "lock")
	if err != nil {
		return nil, err
	}

	unlockNode, err := genNode(name, "unlock")
	if err != nil {
		return nil, err
	}

	lock := &Lock{
		LockID:     NewLockID(name),
		LockNode:   lockNode,
		UnlockID:   NewUnlockID(name),
		UnlockNode: unlockNode,
	}

	return lock, nil
}

// IsLockNode returns true if the node is a lock node
func IsLockNode(n *parse.Node) bool {
	if len(n.Keys) < 0 {
		return false
	}
	return n.Kind() == keywords["lock"]
}

// IsUnlockNode returns true if the node is an unlock node
func IsUnlockNode(n *parse.Node) bool {
	if len(n.Keys) < 0 {
		return false
	}
	return n.Kind() == keywords["unlock"]
}

// NewLockID returns an id of the lock node based on the name
func NewLockID(name string) string {
	return fmt.Sprintf("%s.%s", keywords["lock"], name)
}

// NewUnlockID returns an id of the unlock node based on the name
func NewUnlockID(name string) string {
	return fmt.Sprintf("%s.%s", keywords["unlock"], name)
}

// GetLockName returns the name of the lock from the ID
func GetLockName(id string) string {
	var replaceStrings []string
	for _, kw := range keywords {
		replaceStrings = append(replaceStrings, kw+".", "")
	}
	replacer := strings.NewReplacer(replaceStrings...)
	return replacer.Replace(id)
}

// HasLock returns true if the node has a lock on it
func HasLock(n *parse.Node) (bool, error) {
	name, err := getLock(n)
	if err != nil {
		return false, err
	}
	return name != "", nil
}

// GetLockKeyword returns the name of the lock entry node resource
func GetLockKeyword() string {
	return keywords["lock"]
}

// GetUnlockKeyword returns the name of the lock entry node resource
func GetUnlockKeyword() string {
	return keywords["unlock"]
}

func getLock(n *parse.Node) (string, error) {
	name, err := n.GetString("lock")
	if err != nil && err != parse.ErrNotFound {
		return "", err
	}
	return name, nil
}
