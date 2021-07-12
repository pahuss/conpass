package main

import (
	"conpass/encoder"
	"conpass/helpers"
	"conpass/stores/file"
	"conpass/util"
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
	WorkDir = ".conpass"
)

type Store interface {
	Get(key string) ([]byte, error)
	Add(key string, data []byte) error
	Edit(key string, data []byte) error
	Delete(key string) error
	SetEncodeKey(key, salt string)
}

func newStore(args ...interface{}) Store {
	return &file.Store{
		WorkDir: args[0].(string),
		Encoder: args[1].(encoder.Encoder),
	}
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
		fmt.Println("You need to set a password to encrypt your data. If you forget it or lose\n" +
			"your data will be lost.")
		fmt.Print("Set your password: ")
		bytePassword, _ := term.ReadPassword(syscall.Stdin)
		err := util.SetPassword(dataDirPath, string(bytePassword), salt)
		fmt.Println()

		if err != nil {
			fmt.Println("error setting password")
			os.Exit(1)
		}

		fmt.Print("Confirm password: ")
		confirmBytePassword, _ := term.ReadPassword(syscall.Stdin)
		fmt.Println()

		if string(confirmBytePassword) != string(bytePassword) {
			fmt.Println("error setting password")
			os.Exit(1)
		}

		err = util.SetPassword(dataDirPath, string(bytePassword), salt)
		if err != nil {
			fmt.Println("error setting password")
			os.Exit(1)
		}

		fmt.Println("Success")
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
			fmt.Print("Password: ")
			bytePassword, _ := term.ReadPassword(syscall.Stdin)
			fmt.Println()
			if !util.CheckPassword(dataDirPath, string(bytePassword), salt) {
				printError(errors.New("wrong password\n"))
			}
			store.SetEncodeKey(string(bytePassword), salt)
			GetAction(store, flags["name"].Value.(string), flags["verbose"].Value.(bool))
		})

	// configure info command
	commando.
		Register("add").
		SetShortDescription("").
		SetDescription("").
		AddFlag("name,n", "Data name", commando.String, "").
		AddFlag("data,d", "Data body", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Print("Password: ")
			bytePassword, _ := term.ReadPassword(syscall.Stdin)
			fmt.Println()
			if !util.CheckPassword(dataDirPath, string(bytePassword), salt) {
				printError(errors.New("wrong password\n"))
			}
			store.SetEncodeKey(string(bytePassword), salt)
			name, data := getFlags(flags)
			AddAction(store, name, data)
		})

	// parse command-line arguments
	commando.Parse(nil)
}

func AddAction(store Store, name, data string) {
	if name == "" {
		printError(errors.New("data name must be not empty string"))
	}
	if data == "" {
		printError(errors.New("data must be not empty string"))
	}

	err := store.Add(name, []byte(data))

	if err != nil {
		printError(err)
	} else {
		printSuccess()
	}
}

func GetAction(store Store, key string, verbose bool) {
	data, err := store.Get(key)

	if err != nil {
		printError(err)
	}

	if !verbose {
		err := clipboard.WriteAll(string(data))
		if err != nil {
			printError(err)
		}
		fmt.Println("Copied to clipboard...")
	} else {
		fmt.Println(string(data))
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

func getFlags(args map[string]commando.FlagValue) (string, string) {
	return args["name"].Value.(string), args["data"].Value.(string)
}
