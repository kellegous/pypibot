package rpc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"pypibot/store"
)

func createAndOpenStore(path string) (*store.Store, error) {
	if err := store.Create(path); err != nil {
		return nil, err
	}

	return store.Open(path)
}

func TestConnect(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	s, err := createAndOpenStore(filepath.Join(tmp, "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	c, err := Serve(s)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// TODO(knorton): Connect the client
}
