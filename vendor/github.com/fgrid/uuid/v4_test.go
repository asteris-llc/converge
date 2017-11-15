package uuid

import "testing"

func TestNewV4(t *testing.T) {
	uuid := NewV4()
	if uuid.Version() != 4 {
		t.Errorf("invalid version %d - expected 4", uuid.Version())
	}
	t.Logf("UUID V4: %s", uuid)
}

func BenchmarkNewV4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewV4()
	}
}
