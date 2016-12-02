package names

import "github.com/asteris-llc/converge/parse"

// Fuzz name validation
func Fuzz(data []byte) int {
	err := parse.ValidateName(string(data))

	if err != nil {
		return 1
	}
	return 0
}
