package store

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/scalingdata/gcfg"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"pypibot/auth"
)

const (
	configFilePath = "config.gcfg"
	userFilePath   = "user.db"

	defaultWebAddr = ":8080"
	defaultRpcAddr = ":8081"

	bitsInRsaKeys = 2048

	srvCrtFile = "srv.crt.pem"
	srvKeyFile = "srv.key.pem"

	ServerName = "kellego.us"
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

	db   *leveldb.DB
	path string
}

func (s *Store) Close() error {
	return s.db.Close()
}

func newCAPool(crtPem *pem.Block) (*x509.CertPool, error) {
	p := x509.NewCertPool()

	crt, err := x509.ParseCertificate(crtPem.Bytes)
	if err != nil {
		return nil, err
	}

	p.AddCert(crt)

	return p, nil
}

func (s *Store) ServerTlsConfig() (*tls.Config, error) {
	crtPem, keyPem, err := auth.ReadBothPems(
		filepath.Join(s.path, srvCrtFile),
		filepath.Join(s.path, srvKeyFile))
	if err != nil {
		return nil, err
	}

	prv, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return nil, err
	}

	crt := tls.Certificate{
		Certificate: [][]byte{crtPem.Bytes},
		PrivateKey:  prv,
	}

	p, err := newCAPool(crtPem)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    p,
	}, nil
}

// CreateUser ...
func (s *Store) CreateUser(email, name string, t User_UserType) (*User, *pem.Block, *pem.Block, error) {
	srvCrtPem, srvKeyPem, err := auth.ReadBothPems(
		filepath.Join(s.path, srvCrtFile),
		filepath.Join(s.path, srvKeyFile))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Unable to read server cert: %s", err)
	}

	crtPem, keyPem, err := auth.GenerateClientCert(bitsInRsaKeys, srvCrtPem, srvKeyPem)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to generate client cert: %s", err)
	}

	user := &User{
		Email: email,
		Name:  name,
		Type:  t,
	}

	if err := s.AddUser(user, keyPem); err != nil {
		return nil, nil, nil, fmt.Errorf("unable to insert user: %s", err)
	}

	return user, crtPem, keyPem, nil
}

func (s *Store) AddUser(user *User, key *pem.Block) error {
	return addUser(s.db, user, key)
}

func (s *Store) FindUser(key []byte) (*User, error) {
	var ro opt.ReadOptions

	val, err := s.db.Get(key, &ro)
	if err != nil {
		return nil, err
	}

	var user User
	if err := proto.Unmarshal(val, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Store) ForEachUser(f func([]byte, *User) error) error {
	var ro opt.ReadOptions
	it := s.db.NewIterator(nil, &ro)
	defer it.Release()

	var user User
	for it.Next() {
		if err := proto.Unmarshal(it.Value(), &user); err != nil {
			return err
		}

		if err := f(it.Key(), &user); err != nil {
			return err
		}
	}

	return it.Error()
}

func addUserWithKeyBytes(db *leveldb.DB, user *User, key []byte) error {
	val, err := proto.Marshal(user)
	if err != nil {
		return err
	}

	return db.Put(key, val, &opt.WriteOptions{
		Sync: true,
	})
}

func addUser(db *leveldb.DB, user *User, keyPem *pem.Block) error {
	prv, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return err
	}

	pub, err := x509.MarshalPKIXPublicKey(&prv.PublicKey)
	if err != nil {
		return err
	}

	return addUserWithKeyBytes(db, user, pub)
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

	srvCrt, srvKey, err := auth.GenerateServerCert(bitsInRsaKeys, ServerName)
	if err != nil {
		return err
	}

	if err := auth.WriteBothPems(
		srvCrt,
		filepath.Join(path, srvCrtFile),
		srvKey,
		filepath.Join(path, srvKeyFile)); err != nil {
		return err
	}

	db, err := leveldb.OpenFile(filepath.Join(path, userFilePath), &opt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()

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

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return &Store{
		Config: cfg,
		db:     db,
		path:   abs,
	}, nil
}
