package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/scalingdata/gcfg"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"pypibot/auth"
	"pypibot/pb"
)

const (
	configFilePath = "config.gcfg"
	userFilePath   = "user.db"

	defaultWebAddr = ":8080"
	defaultRpcAddr = ":8081"

	godEmail = "kel@kellegous.com"
	godName  = "God"
)

type Config struct {
	Web struct {
		Addr string
	}

	Rpc struct {
		Addr string
	}
}

func (c *Config) ReadFromFile(filename string) error {
	return gcfg.ReadFileInto(c, filename)
}

type Store struct {
	Config *Config

	db *leveldb.DB
}

func (s *Store) Close() error {
	return s.db.Close()
}

func addUserWithKey(db *leveldb.DB, user *pb.User, key []byte) error {
	val, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	return db.Put(key, val, &opt.WriteOptions{
		Sync: true,
	})
}

func addUserWithKeyFromFile(db *leveldb.DB, user *pb.User, filename string) error {
	prv, err := auth.ReadPrivateKey(filename)
	if err != nil {
		return err
	}

	pub, err := auth.GetPublicKey(prv)
	if err != nil {
		return err
	}

	return addUserWithKey(db, user, pub)
}

func writeDefaultConfig(filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := fmt.Fprintf(w, "[web]\naddr=%s\n\n", defaultWebAddr); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "[rpc]\naddr=%s\n\n", defaultRpcAddr); err != nil {
		return err
	}

	return nil
}

func Create(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists.", path)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	if err := writeDefaultConfig(filepath.Join(path, configFilePath)); err != nil {
		return err
	}

	caCrt := filepath.Join(path, "ca.crt")
	caKey := filepath.Join(path, "ca.key")
	if err := auth.GenerateCa("kellego.us", caCrt, caKey); err != nil {
		return err
	}

	godCrt := filepath.Join(path, "god.crt")
	godKey := filepath.Join(path, "god.key")
	if _, err := auth.GenerateCert(
		"kellego.us",
		caCrt,
		caKey,
		godCrt,
		godKey); err != nil {
		return err
	}

	db, err := leveldb.OpenFile(filepath.Join(path, userFilePath), &opt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

	n := godName
	e := godEmail
	t := pb.User_GOD

	if err := addUserWithKeyFromFile(
		db,
		&pb.User{Name: &n, Email: &e, Type: &t},
		godKey); err != nil {
		return err
	}

	return nil
}

func Open(path string) (*Store, error) {
	cfg := &Config{}

	if err := cfg.ReadFromFile(filepath.Join(path, configFilePath)); err != nil {
		return nil, err
	}

	var o opt.Options
	db, err := leveldb.OpenFile(filepath.Join(path, userFilePath), &o)
	if err != nil {
		return nil, err
	}

	return &Store{
		Config: cfg,
		db:     db,
	}, nil
}
