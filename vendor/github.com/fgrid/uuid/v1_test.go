package uuid

import "testing"

func TestNewV1(t *testing.T) {
	uuid := NewV1()
	if uuid.Version() != 1 {
		t.Errorf("invalid version %d - expected 1", uuid.Version())
	}
	t.Logf("UUID V1: %s", uuid)
}

func BenchmarkNewV1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewV1()
	}
}
