package gin

import (
	"testing"
)

func TestGin(t *testing.T) {
	g, err := NewGin(`
      name: "GinServer"
      addr: ":8080"
      routers:
        GET: /send/article`)
	if err != nil {
		t.Fatal(err)
	}
	err = g.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = g.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
