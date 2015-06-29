package rpc

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"pypibot/auth"
	"pypibot/pb"
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

	data := filepath.Join(tmp, "data")

	s, err := createAndOpenStore(data)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, crtPem, keyPem, err := s.CreateUser("foo@email.com", "foo", pb.User_PERSON)
	if err != nil {
		t.Fatal(err)
	}

	srv, err := Serve(s)
	if err != nil {
		t.Fatal(err)
	}

	srvCrtPem, err := auth.ReadPem(filepath.Join(data, "srv.crt.pem"))
	if err != nil {
		t.Fatal(err)
	}

	clt, err := Dial(":8081", srvCrtPem, crtPem, keyPem)
	if err != nil {
		t.Fatal(err)
	}

	if err := clt.Close(); err != nil {
		t.Fatal(err)
	}

	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
