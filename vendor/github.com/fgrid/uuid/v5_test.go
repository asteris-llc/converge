package uuid

import (
	"fmt"
	"testing"
)

func TestNewV5(t *testing.T) {
	namespace := NewNamespaceUUID("test")
	uuid := NewV5(namespace, []byte("test name"))
	if uuid.Version() != 5 {
		t.Errorf("invalid version %d - expected 5", uuid.Version())
	}
	t.Logf("UUID V5: %s", uuid)
}

func BenchmarkNewV5(b *testing.B) {
	test := NewNamespaceUUID("test")
	for i := 0; i < b.N; i++ {
		NewV5(test, []byte("example"))
	}
}

func ExampleNewNamespaceUUID() {
	fmt.Printf("UUID(test):        %s\n", NewNamespaceUUID("test"))
	fmt.Printf("UUID(myNameSpace): %s\n", NewNamespaceUUID("myNameSpace"))
	// Output:
	// UUID(test):        e8b764da-5fe5-51ed-8af8-c5c6eca28d7a
	// UUID(myNameSpace): 40e41e4d-01d6-5e36-8c6b-93edcdf1442d
	//
}
