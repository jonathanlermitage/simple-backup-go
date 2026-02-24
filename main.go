package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	var password string
	var targetFolder string
	var configFile string
	var logsFolder string
	var workFolder string
	var taskNamesToExecute string
	var restartOneDrive bool
	flag.StringVar(&password, "password", "", "protect archive with given password")
	flag.StringVar(&targetFolder, "targetFolder", "", "archive location")
	flag.StringVar(&configFile, "configFile", "", "configuration file")
	flag.StringVar(&logsFolder, "logsFolder", "", "logs folder location")
	flag.StringVar(&workFolder, "workFolder", "", "temporary work directory for archive generation")
	flag.StringVar(&taskNamesToExecute, "taskNamesToExecute", "", "list of backup tasks to execute")
	flag.BoolVar(&restartOneDrive, "restartOneDrive", false, "restart MS OneDrive once backup completes")
	flag.Parse()

	// Stop OneDrive, similar to %LOCALAPPDATA%\Microsoft\OneDrive\OneDrive.exe /shutdown
	if restartOneDrive && runtime.GOOS == "windows" {
		var onedrivePath = os.Getenv("LOCALAPPDATA") + "\\Microsoft\\OneDrive\\OneDrive.exe"
		fmt.Printf("💤 Stopping OneDrive\n")
		_, err := exec.Command(onedrivePath, "/shutdown").Output()
		if err != nil {
			fmt.Printf("Error stopping OneDrive: %v\n", err)
		}
	}

	var configs = LoadBackupTaskConfigs(configFile)

	if len(logsFolder) > 0 {
		logsFolder = strings.Replace(logsFolder, "\\", "/", -1)
		if !(strings.HasSuffix(logsFolder, "/")) {
			logsFolder += "/"
		}
		RotateLogs(logsFolder)
	}

	var taskNamesToKeepArray []string
	if len(taskNamesToExecute) > 0 {
		taskNamesToKeepArray = strings.Split(taskNamesToExecute, ",")
	} else {
		taskNamesToKeepArray = make([]string, 0)
	}

	var failedTasks list.List
	for i := 0; i < len(configs); i++ {
		var config = configs[i]

		fmt.Print("\n----------[ executing step " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(configs)) + ", backup of " + config.TaskName + " ]----------\n")

		if len(taskNamesToExecute) > 0 && !IsElementExist(taskNamesToKeepArray, config.TaskName) {
			fmt.Println("✔ Skipped by backup tasks filter (taskNamesToExecute)")
			continue
		}

		var archiveName = strings.Replace(config.Source, "\\", "_", -1)
		archiveName = strings.Replace(archiveName, "/", "_", -1)
		archiveName = strings.Replace(archiveName, ":", "", -1)

		compressResult := Compress(config.Source, config.Excludes, targetFolder, archiveName, password, config.ProtectWithPassword == "true", workFolder)

		if len(logsFolder) > 0 {
			SaveLogs(compressResult, logsFolder)
		}

		if validate7zOutput(compressResult) {
			fmt.Println("✅ Compression completed with success")

		} else {
			failedTasks.PushFront("❌ " + config.TaskName + ":\nBad compression output: " + compressResult)
		}
	}

	// Start OneDrive in the background, similar to start "OneDrive" /B "%LOCALAPPDATA%\Microsoft\OneDrive\onedrive" /background
	if restartOneDrive && runtime.GOOS == "windows" {
		var onedrivePath = os.Getenv("LOCALAPPDATA") + "\\Microsoft\\OneDrive\\OneDrive.exe"
		fmt.Printf("\n🚀 Starting OneDrive")
		cmd := exec.Command(onedrivePath, "/background")
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting OneDrive: %v\n", err)
			return
		}
	}

	if failedTasks.Len() > 0 {
		fmt.Println("\n----------[ ❌ error report ❌ ]----------")
		SaveLogs(strconv.Itoa(failedTasks.Len())+" task(s) failed", logsFolder)
		for e := failedTasks.Front(); e != nil; e = e.Next() {
			SaveLogs(fmt.Sprint(e.Value), logsFolder)
			fmt.Println(e.Value)
		}
		fmt.Println("\n" + strconv.Itoa(failedTasks.Len()) + " task(s) failed")
	} else {
		fmt.Println("\n😎 Everything is OK! 😎")
		SaveLogs("Everything is OK!", logsFolder)
	}

	fmt.Print("\nPress any key to exit...")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	fmt.Println(input.Text())
}
