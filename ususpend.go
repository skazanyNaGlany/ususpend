package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

const AppName = "USUSPEND"
const AppVersion = "0.1"
const IgnoreFilePathname = "./ususpend.ignore.txt"
const MinUid = 1000 // users UIDs start from 1000
const defaultIgnore = `# processes to be ignored, by command line, line by line
#
# lines started with # will be ignored
# you can use regex
#
# example of ignored process by full command line
# /opt/google/chrome/chrome --type=renderer --crashpad-handler-pid=.* --enable-crash-reporter=.*

# do not touch ususpend
.*ususpend.*

# do not touch docker
.*docker.*
`

var exeDir = filepath.Dir(os.Args[0])
var fullIgnoreFilePathname = filepath.Join(exeDir, IgnoreFilePathname)
var ignoreData = make([]*regexp.Regexp, 0)

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
	log.Println("Suspend or resume non-system (users) processes.")
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

	log.Printf("%v does not exists, creating default.", fullIgnoreFilePathname)

	os.WriteFile(fullIgnoreFilePathname, []byte(defaultIgnore), 0666)

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

		compiledRegEx, err := regexp.Compile(line)

		if err != nil {
			log.Fatalln("cannot compile", line, ":", err)
		}

		ignoreData = append(ignoreData, compiledRegEx)
	}
}

func changeCurrentWorkingDir() {
	os.Chdir(exeDir)
}

func resume(resume bool) {
	processes, err := process.Processes()

	if err != nil {
		log.Fatalln(err)
	}

	for _, iprocess := range processes {
		cmdLine, err := iprocess.Cmdline()

		if err != nil {
			log.Println(err)

			continue
		}

		uids, err := iprocess.Uids()

		if err != nil {
			log.Println(err)

			continue
		}

		cmdLine = strings.TrimSpace(cmdLine)

		if !isUserProcess(uids) {
			log.Println("ignore", cmdLine, "[system]")

			continue
		}

		if isIgnoredProcess(cmdLine) {
			log.Println("ignore", cmdLine)

			continue
		}

		if resume {
			log.Println("resume", cmdLine)

			err = iprocess.SendSignal(syscall.SIGCONT)
		} else {
			log.Println("suspend", cmdLine)

			err = iprocess.SendSignal(syscall.SIGSTOP)
		}

		if err != nil {
			log.Println("cannot send signal to", cmdLine, ":", err)
		}
	}
}

func isIgnoredProcess(cmdLine string) bool {
	for _, rex := range ignoreData {
		if rex.MatchString(cmdLine) {
			return true
		}
	}

	return false
}

func isUserProcess(uids []int32) bool {
	for _, iuid := range uids {
		if iuid >= MinUid {
			return true
		}
	}

	return false
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
		resume(true)
	} else if os.Args[1] == "--suspend" {
		createIgnoreFile()
		readIgnoreFile()
		resume(false)
	} else {
		printAppInfo()
		printUsages()
	}
}
