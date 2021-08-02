package main

import (
	"conpass/encoder"
	"conpass/helpers"
	"conpass/stores/file"
	"conpass/util"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/thatisuday/commando"
	"golang.org/x/term"
	"os"
	"os/user"
	"path"
	"syscall"
)

const (
	WorkDir                      = ".conpass"
	codeErrorWhenSetPassword     = 0
	codeUnEqualConfirmedPassword = 1
	loginPasswordStrategy        = 1
	stringStrategy               = 2
)

type setPasswordError struct {
	code int
	s    string
}

func (e *setPasswordError) Error() string {
	return e.s
}

func unequalPasswordsError() *setPasswordError {
	return &setPasswordError{codeUnEqualConfirmedPassword, "confirmed password must be equal source"}
}

func isUnequalPasswordsError(e *setPasswordError) bool {
	return e.code == codeUnEqualConfirmedPassword
}

func whenSetPasswordError() *setPasswordError {
	return &setPasswordError{codeErrorWhenSetPassword, "error when password set"}
}

func isWhenSetPasswordError(e *setPasswordError) bool {
	return e.code == codeErrorWhenSetPassword
}

type Store interface {
	Get(key string) ([]byte, error)
	Add(key string, data []byte, args ...interface{}) error
	Edit(key string, data []byte, args ...interface{}) error
	Delete(key string) error
	SetEncodeKey(key, salt string)
}

func newStore(args ...interface{}) Store {
	return &file.Store{
		WorkDir: args[0].(string),
		Encoder: args[1].(encoder.Encoder),
	}
}

type Data struct {
	data     interface{}
	strategy int
}
type loginPasswordData struct {
	login    string
	password string
}

func main() {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	workDir := path.Join(homeDir, WorkDir)
	dirInfo, err := os.Stat(workDir)

	if err != nil {
		createErr := os.Mkdir(workDir, 0700)
		if createErr != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if !dirInfo.IsDir() {
		fmt.Printf("%s if file but need directory with same name", workDir)
		os.Exit(1)
	}

	current, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	salt := helpers.GetMD5Hash(current.Username)
	dataDirPath := path.Join(workDir, salt)

	dataDirInfo, err := os.Stat(dataDirPath)

	if err != nil {
		createErr := os.Mkdir(dataDirPath, 0700)
		if createErr != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if !dataDirInfo.IsDir() {
		fmt.Printf("%s data dir is file\n", dataDirInfo)
		os.Exit(1)
	}

	if !util.CheckIsPasswordWasSet(dataDirPath) {
		err := setPassword(dataDirPath, salt)
		if err != nil {
			if isUnequalPasswordsError(err.(*setPasswordError)) {
				fmt.Println("Confirmed password must be equal source password. Exit")
			} else {
				fmt.Println(err.Error() + " Exit")
			}
			os.Exit(1)
		}
	}

	store := newStore(dataDirPath, encoder.NewEncoder())

	// configure commando
	commando.
		SetExecutableName("conpass").
		SetVersion("1.0.0").
		SetDescription("Console password manager")

	// configure the root command
	commando.
		Register(nil).
		AddFlag("name,n", "Data name", commando.String, "").
		AddFlag("verbose,V", "Out data to screen", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			password, err := askPassword(dataDirPath, salt)
			if err != nil {
				printError(err)
			}
			store.SetEncodeKey(password, salt)
			GetAction(store, flags["name"].Value.(string), flags["verbose"].Value.(bool))
		})

	// configure info command
	commando.
		Register("add").
		SetShortDescription("").
		SetDescription("").
		AddFlag("name,n", "Resource that sets credentials name", commando.String, "").
		AddFlag("login,l", "Resource login", commando.String, "").
		AddFlag("password,p", "Resource password", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			password, err := askPassword(dataDirPath, salt)
			if err != nil {
				printError(err)
			}
			store.SetEncodeKey(password, salt)
			resourceName, resourceLogin, resourcePassword := getFlags(flags)
			AddAction(store, resourceName, resourceLogin, resourcePassword)
		})

	// parse command-line arguments
	commando.Parse(nil)
}

func askPassword(dataDirPath, passwordSalt string) (string, error) {
	fmt.Print("Password: ")
	bytePassword, _ := term.ReadPassword(syscall.Stdin)
	fmt.Println()
	if !util.CheckPassword(dataDirPath, string(bytePassword), passwordSalt) {
		return "", errors.New("wrong password\n")
	}
	return string(bytePassword), nil
}

func setPassword(dataDirPath, passwordSalt string) error {
	fmt.Println("You need to set a password to encrypt your data. If you forget it or lose\n" +
		"your data will be lost.")
	fmt.Print("Set your password: ")
	bytePassword, _ := term.ReadPassword(syscall.Stdin)
	err := util.SetPassword(dataDirPath, string(bytePassword), passwordSalt)
	fmt.Println()

	if err != nil {
		return err
	}

	fmt.Print("Confirm password: ")
	confirmBytePassword, _ := term.ReadPassword(syscall.Stdin)
	fmt.Println()

	if string(confirmBytePassword) != string(bytePassword) {
		return unequalPasswordsError()
	}

	err = util.SetPassword(dataDirPath, string(bytePassword), passwordSalt)
	if err != nil {
		return whenSetPasswordError()
	}

	fmt.Println("Success")
	return nil
}

func AddAction(store Store, name, login, password string) {
	if name == "" && login == "" {
		printError(errors.New(" name or login must be not empty string"))
	}
	if login == "" {
		printError(errors.New("data must be not empty string"))
	}
	openData := Data{loginPasswordData{login, password}, loginPasswordStrategy}

	d, err := json.Marshal(openData)
	if err != nil {
		printError(errors.New("something went wrong"))
	}

	err = store.Add(name, d)

	if err != nil {
		printError(err)
	} else {
		printSuccess()
	}
}

func GetAction(store Store, key string, verbose bool) {
	jsonData, err := store.Get(key)

	if err != nil {
		printError(err)
	}

	d := Data{}

	err = json.Unmarshal(jsonData, &d)

	if !verbose {
		err := clipboard.WriteAll(d.data.(loginPasswordData).password)
		if err != nil {
			printError(err)
		}
		fmt.Println("Copied to clipboard...")
	} else {
		fmt.Println(d.data.(loginPasswordData).password)
	}
}

func printError(err error) {
	fmt.Println(err.Error())
	os.Exit(0)
}

func printSuccess() {
	fmt.Println("Success")
	os.Exit(0)
}

func getFlags(args map[string]commando.FlagValue) (string, string, string) {
	return args["name"].Value.(string), args["login"].Value.(string), args["password"].Value.(string)
}
