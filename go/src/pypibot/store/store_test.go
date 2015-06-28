package store

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"pypibot/pb"
)

func getUserCount(s *Store) (int, error) {
	c := 0
	if err := s.ForEachUser(func(key []byte, user *pb.User) error {
		c++
		return nil
	}); err != nil {
		return 0, err
	}
	return c, nil
}

func TestCreateAndOpen(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	dst := filepath.Join(tmp, "data")

	if err := Create(dst); err != nil {
		t.Fatal(err)
	}

	s, err := Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	tc, err := s.ServerTlsConfig()
	if err != nil {
		t.Fatal(err)
	}

	if len(tc.Certificates) == 0 {
		t.Fatalf("Expected tls.Config with >0 certs, got %d certs",
			len(tc.Certificates))
	}

	if s.Config.Rpc.Addr != defaultRpcAddr {
		t.Fatalf("config's rpc addr should be %s, got %s",
			s.Config.Rpc.Addr,
			defaultRpcAddr)
	}

	if s.Config.Web.Addr != defaultWebAddr {
		t.Fatalf("config's web addr should be %s, got %s",
			s.Config.Web.Addr,
			defaultWebAddr)
	}

	uc, err := getUserCount(s)
	if err != nil {
		t.Fatal(err)
	}

	if uc != 1 {
		t.Fatalf("expected only 1 user, got %d", uc)
	}
}
