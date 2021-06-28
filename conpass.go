package main

import (
	"bufio"
	"conpass/helpers"
	"conpass/util"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/thatisuday/commando"
	"os"
	"os/user"
	"path"
)

const (
	WorkDir = ".conpass"
	//Salt = "a2de4fs"
)

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

	dataDirPath := path.Join(workDir, helpers.GetMD5Hash(current.Username))

	dataDirInfo, err := os.Stat(dataDirPath)

	if err != nil {
		createErr := os.Mkdir(dataDirPath, 0700)
		if createErr != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if !dataDirInfo.IsDir() {
		fmt.Printf("%s data dir is file", dataDirInfo)
		os.Exit(1)
	}

	// configure commando
	commando.
		SetExecutableName("conpass").
		SetVersion("1.0.0").
		SetDescription("Console password manager")

	// configure the root command
	commando.
		Register(nil).
		AddFlag("name,n", "Data name", commando.String, "").
		AddFlag("buffer,b", "Copy to clipboard", commando.Bool, true).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			GetAction(flags["name"].Value.(string), dataDirPath, flags["buffer"].Value.(bool))
		})

	// configure info command
	commando.
		Register("add").
		SetShortDescription("").
		SetDescription("").
		AddFlag("name,n", "Data name", commando.String, "").
		AddFlag("data,d", "Data body", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			name, data := getFlags(flags)
			AddAction(name, data, dataDirPath)
		})

	// parse command-line arguments
	commando.Parse(nil)
}

func AddAction(name, data, workDir string) {
	file, err := os.OpenFile(path.Join(workDir, helpers.GetMD5Hash(name+"")), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		printError(err.Error())
	}
	err = util.ActionAdd(name, data, file)

	if err != nil {
		printError(err.Error())
	} else {
		printSuccess()
	}
}

func GetAction(name, workDir string, copyToClipboard bool) {
	fileName := path.Join(workDir, helpers.GetMD5Hash(name+""))
	info, err := os.Stat(fileName)

	if err != nil {
		printError("Not found")
	}

	file, err := os.OpenFile(path.Join(workDir, helpers.GetMD5Hash(name+"")), os.O_RDONLY, 0600)
	if err != nil {
		printError(err.Error())
	}

	r := bufio.NewReader(file)
	data, err := r.Peek(int(info.Size()))

	if err != nil {
		printError(err.Error())
	}

	if copyToClipboard {
		err := clipboard.WriteAll(string(data))
		if err != nil {
			printError(err.Error())
		}
		fmt.Println("Copied to clipboard...")
	} else {
		fmt.Println(string(data))
	}
}

func printError(message string) {
	fmt.Println(message)
	os.Exit(0)
}

func printSuccess() {
	fmt.Println("Success")
	os.Exit(0)
}

func getFlags(args map[string]commando.FlagValue) (string, string) {
	return args["name"].Value.(string), args["data"].Value.(string)
}
