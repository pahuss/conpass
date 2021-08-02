package file

import (
	"bufio"
	"conpass/encoder"
	"conpass/helpers"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"os"
	"path"
)

type Store struct {
	WorkDir string
	Encoder encoder.Encoder
}

func (s *Store) Get(key string) ([]byte, error) {
	fileName := path.Join(s.WorkDir, helpers.GetMD5Hash(key+""))
	info, err := os.Stat(fileName)

	if err != nil {
		return nil, errors.New("not found")
	}

	file, err := os.OpenFile(path.Join(s.WorkDir, helpers.GetMD5Hash(key)), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	data, err := r.Peek(int(info.Size()))
	if err != nil {
		return nil, err
	}

	data, err = s.Encoder.Decrypt(data)
	if err != nil {
		return nil, errors.New("decryption error")
	}
	return data, nil
}

func (s *Store) Add(key string, data []byte, args ...interface{}) error {
	openFile, err := os.OpenFile(path.Join(s.WorkDir, helpers.GetMD5Hash(key)), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer openFile.Close()
	data, err = s.Encoder.Encrypt(data)
	if err != nil {
		return errors.New("encryption error")
	}
	_, err = openFile.Write(data)
	if err != nil {
		return err
	}
	err = s.AddToRegistry(key)
	return err
}

func (s *Store) Edit(key string, data []byte, args ...interface{}) error {
	return s.Add(key, data)
}
func (s *Store) Delete(key string) error {
	return nil
}

func (s *Store) SetEncodeKey(key, salt string) {
	h := sha1.New()
	h.Write([]byte((key + salt)))
	s.Encoder.SetKey([]byte(helpers.GetMD5Hash(hex.EncodeToString(h.Sum(nil)))))
}

func (s *Store) AddToRegistry(key string) error {
	registryData, err := s.DecryptRegistry()
	if err != nil {
		return err
	}
	registryData = s.AddToRegistryData(registryData, key)
	err = s.EncryptRegistry(registryData)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) EncryptRegistry(registryData []byte) error {
	registryFilePath := path.Join(s.WorkDir, ".registry")
	openFile, err := os.OpenFile(registryFilePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer openFile.Close()
	data, err := s.Encoder.Encrypt(registryData)
	if err != nil {
		return errors.New("registry encryption error")
	}
	_, err = openFile.Write(data)
	return err
}

func (s *Store) DecryptRegistry() ([]byte, error) {
	registryFilePath := path.Join(s.WorkDir, ".registry")
	file, err := os.OpenFile(registryFilePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	info, err := os.Stat(registryFilePath)
	if err != nil {
		return nil, err
	}
	data, err := r.Peek(int(info.Size()))
	if err != nil {
		return nil, err
	}

	data, err = s.Encoder.Decrypt(data)
	if err != nil {
		return nil, errors.New("registry decryption error")
	}

	return data, nil
}

func (s *Store) AddToRegistryData(registryData []byte, data string) []byte {
	return []byte(string(registryData) + "," + data)
}
