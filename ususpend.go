package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const AppName = "USUSPEND"
const AppVersion = "0.1"
const IgnoreFilePathname = "./ususpend.ignore"

var exeDir = filepath.Dir(os.Args[0])
var fullIgnoreFilePathname = filepath.Join(exeDir, IgnoreFilePathname)
var ignoreData = make([]string, 0)

func duplicateLog() {
	logFilename := filepath.Base(os.Args[0]) + ".txt"
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
}

func getFullAppName() string {
	return fmt.Sprintf("%v v%v", AppName, AppVersion)
}

func printAppName() {
	log.Println(
		getFullAppName())
	log.Println()
	log.Println("Suspend or resume non-system processes.")
	log.Println()
}

func printAppInfo() {
}

func printUsages() {
	log.Printf("Usage: %v <option>", os.Args[0])

	log.Println()
	log.Println("Options:")

	log.Println("\t--resume")
	log.Println("\t\t\t resume all non-system processes")
	log.Println()
	log.Println("\t--suspend")
	log.Println("\t\t\t suspend all non-system processes")
	log.Println()
}

func shouldPrintUsages() bool {
	len_args := len(os.Args)

	return len_args != 2 || (len_args > 1 && os.Args[1] == "--help")
}

func checkPlatform() {
	if runtime.GOOS != "linux" {
		log.Fatalln("This app can be used only on Linux.")
	}
}

func createIgnoreFile() {
	if _, err := os.Stat(fullIgnoreFilePathname); err == nil {
		return
	}

	empty := `# processes to be ignored, line by line
	# you can use regex, for example:
	# .*docker.*
	`

	log.Printf("%v does not exists, creating default.", fullIgnoreFilePathname)

	os.WriteFile(fullIgnoreFilePathname, []byte(empty), 0666)

	log.Printf("%v created.\n", fullIgnoreFilePathname)
}

func readIgnoreFile() {
	data, err := os.ReadFile(fullIgnoreFilePathname)

	if err != nil {
		log.Fatalln(err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line[0] == '#' {
			continue
		}

		ignoreData = append(ignoreData, line)
	}
}

func changeCurrentWorkingDir() {
	os.Chdir(exeDir)
}

func main() {
	changeCurrentWorkingDir()
	duplicateLog()
	printAppName()
	checkPlatform()

	if shouldPrintUsages() {
		printAppInfo()
		printUsages()

		os.Exit(1)
	}

	if os.Args[1] == "--resume" {
		createIgnoreFile()
		readIgnoreFile()
	} else if os.Args[1] == "--suspend" {
		createIgnoreFile()
		readIgnoreFile()
	} else {
		printAppInfo()
		printUsages()
	}
}
