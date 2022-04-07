package util

import (
	"github.com/pahuss/conpass/helpers"
	"io/ioutil"
	"os"
	"path"
)

const passwordHashFileName = ".h"

func SetPassword(workdir, password, salt string) error {
	file, err := os.OpenFile(path.Join(workdir, passwordHashFileName), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(helpers.GetMD5Hash(concat(password, salt))))
	if err != nil {
		return err
	}
	return nil
}

func CheckPassword(workdir, password, salt string) bool {
	data, err := ioutil.ReadFile(path.Join(workdir, passwordHashFileName))
	if err != nil {
		return false
	}
	return string(data) == helpers.GetMD5Hash(concat(password, salt))
}

func CheckIsPasswordWasSet(workdir string) bool {
	fileInfo, err := os.Stat(path.Join(workdir, passwordHashFileName))
	return err == nil && !fileInfo.IsDir()
}

func concat(a, b string) string {
	return a + b
}
