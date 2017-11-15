package uuid

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	uuid := UUID{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00,
		0x00,
		0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	uuid[6] = 0x10
	if uuid.Version() != 1 {
		t.Errorf("invalid version %d - expected 1", uuid.Version())
	}
	uuid[6] = 0x20
	if uuid.Version() != 2 {
		t.Errorf("invalid version %d - expected 2", uuid.Version())
	}
	uuid[6] = 0x30
	if uuid.Version() != 3 {
		t.Errorf("invalid version %d - expected 3", uuid.Version())
	}
	uuid[6] = 0x40
	if uuid.Version() != 4 {
		t.Errorf("invalid version %d - expected 4", uuid.Version())
	}
	uuid[6] = 0x50
	if uuid.Version() != 5 {
		t.Errorf("invalid version %d - expected 5", uuid.Version())
	}
}

func ExampleString_NIL() {
	fmt.Printf("NIL-UUID: %s", NIL.String())
	// Output:
	// NIL-UUID: 00000000-0000-0000-0000-000000000000
}
