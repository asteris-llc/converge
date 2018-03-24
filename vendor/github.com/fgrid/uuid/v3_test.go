package uuid

import "testing"

func TestNewV3(t *testing.T) {
	namespace := NewNamespaceUUID("test")
	uuid := NewV3(namespace, []byte("test name"))
	if uuid.Version() != 3 {
		t.Errorf("invalid version %d - expected 3", uuid.Version())
	}
	t.Logf("UUID V3: %s", uuid)
}

func BenchmarkNewV3(b *testing.B) {
	test := NewNamespaceUUID("test")
	for i := 0; i < b.N; i++ {
		NewV3(test, []byte("example"))
	}
}
