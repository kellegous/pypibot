package auth

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func GenerateRsaKey(filename string, size int) error {
	if err := exec.Command(
		"openssl",
		"genrsa",
		"-out",
		filename,
		fmt.Sprintf("%d", size)).Run(); err != nil {
		return fmt.Errorf("unable to create rsa key: %s", err)
	}

	return nil
}

func GenerateCa(name, crtFile, keyFile string) error {
	if err := GenerateRsaKey(keyFile, 2048); err != nil {
		return err
	}

	if err := exec.Command(
		"openssl",
		"req",
		"-x509",
		"-new",
		"-key", keyFile,
		"-out", crtFile,
		"-days", "730",
		"-subj", fmt.Sprintf("/CN=\"%s\"", name)).Run(); err != nil {
		return fmt.Errorf("unable to create cert: %s", err)
	}

	return nil
}

func GenerateCert(host, caCrtFile, caKeyFile, crtFile, keyFile string) error {
	if err := GenerateRsaKey(keyFile, 2048); err != nil {
		return err
	}

	reqFile, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer func() {
		reqFile.Close()
		os.Remove(reqFile.Name())
	}()

	if err := exec.Command(
		"openssl",
		"req",
		"-new",
		"-out", reqFile.Name(),
		"-key", keyFile,
		"-subj", fmt.Sprintf("/CN=%s", host)).Run(); err != nil {
		return fmt.Errorf("unable to create signing request: %s", err)
	}

	if err := exec.Command(
		"openssl",
		"x509",
		"-req",
		"-in", reqFile.Name(),
		"-out", crtFile,
		"-CAkey", caKeyFile,
		"-CA", caCrtFile,
		"-days", "365",
		"-CAcreateserial",
		"-CAserial", "serial").Run(); err != nil {
		return fmt.Errorf("unable to create certificate: %s", err)
	}

	return nil
}
