package main

import (
	"container/list"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	var dryRun bool
	var password string
	var targetFolder string
	var configFile string
	var logsFolder string
	var taskNamesToExecute string
	flag.BoolVar(&dryRun, "dryRun", true, "dry run")
	flag.StringVar(&password, "password", "", "protect archive with given password")
	flag.StringVar(&targetFolder, "targetFolder", "", "archive location")
	flag.StringVar(&configFile, "configFile", "", "configuration file")
	flag.StringVar(&logsFolder, "logsFolder", "", "logs folder location")
	flag.StringVar(&taskNamesToExecute, "taskNamesToExecute", "", "list of backup tasks to execute")
	flag.Parse()

	var configs = LoadBackupTaskConfigs(configFile)

	if len(logsFolder) > 0 {
		logsFolder = strings.Replace(logsFolder, "\\", "/", -1)
		if !(strings.HasSuffix(logsFolder, "/")) {
			logsFolder += "/"
		}
		RotateLogs(logsFolder, dryRun)
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

		compressResult := Compress(config.Source, config.Excludes, targetFolder, archiveName, password, config.ProtectWithPassword == "true", dryRun)

		if len(logsFolder) > 0 {
			SaveLogs(compressResult, logsFolder, dryRun)
		}

		if validate7zOutput(compressResult) {
			fmt.Println("✔ Compression done: " + compressResult)

		} else {
			failedTasks.PushFront("♦ " + config.TaskName + ":\nBad compression output: " + compressResult)
		}
	}

	if failedTasks.Len() > 0 {
		fmt.Println("\n----------[ error report ]----------")
		SaveLogs(strconv.Itoa(failedTasks.Len())+" task(s) failed", logsFolder, dryRun)
		for e := failedTasks.Front(); e != nil; e = e.Next() {
			SaveLogs(fmt.Sprint(e.Value), logsFolder, dryRun)
			fmt.Println(e.Value)
		}
		fmt.Println("\n" + strconv.Itoa(failedTasks.Len()) + " task(s) failed")
	} else {
		fmt.Println("\n😎 Everything is OK!")
		SaveLogs("Everything is OK!", logsFolder, dryRun)
	}
}
