package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// FileExists returns true is given file exists, otherwise false.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FolderExists returns true is given folder exists, otherwise false.
func FolderExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetCurrentDateTime gets current date time in "YYYYMMDD hhmmss" format.
func GetCurrentDateTime() string {
	return time.Now().Format("20060102 150405")
}

// GetCurrentDate gets current date time in "YYYYMM" format.
func GetCurrentDate() string {
	return time.Now().Format("200601")
}

// Compress compresses given folder to archive using 7z method, and optionally protects archive
// with a password.
func Compress(folderToCompress string, excludes []string, targetFolder string, archiveName string, password string, protectWithPassword bool, dryRun bool) string {
	args := []string{"a", "-t7z"}

	if protectWithPassword && len(password) > 0 {
		args = append(args, "-p"+password)
	}

	folderToCompress = strings.Replace(folderToCompress, "\\", "/", -1)
	if strings.HasSuffix(folderToCompress, "/") {
		folderToCompress = folderToCompress[0 : len(folderToCompress)-1]
	}

	targetFolder = strings.Replace(targetFolder, "\\", "/", -1)
	if !(strings.HasSuffix(targetFolder, "/")) {
		targetFolder += "/"
	}
	targetFolder += GetCurrentDate() + "/"

	archiveName = targetFolder + archiveName + " " + GetCurrentDateTime() + ".7z"

	if !dryRun {
		if !FolderExists(targetFolder) {
			mkdirErr := os.MkdirAll(targetFolder, 0755)
			if mkdirErr != nil {
				log.Fatal(mkdirErr)
			}
		}
		if FileExists(archiveName) {
			_ = os.Remove(archiveName)
		}
	}

	args = append(args, archiveName, folderToCompress,
		"-ssw", /* compress files open for writing */
		"-mx3", /* set level of compression */
		"-bd" /* disable progress indicator */)

	var firstFolder = folderToCompress[strings.LastIndex(folderToCompress, "/")+1:]
	for i := 0; i < len(excludes); i++ {
		args = append(args, "-xr!"+firstFolder+"/"+excludes[i])
	}

	fmt.Println("Will run 7z with args:", args)
	if dryRun {
		return "Everything is Ok"
	}

	out, err := exec.Command("7z.exe", args...).Output()

	if err != nil {
		fmt.Println(err)
	}
	return string(out[:])
}

func RotateReport(logsFolder string, dryRun bool) {
	var prevReportFilePath = logsFolder + "simple_backup-go-report_prev.txt"
	var reportFilePath = logsFolder + "simple_backup-go-report.txt"
	fmt.Println("Report rotation: move " + reportFilePath + " to " + prevReportFilePath)
	if !dryRun {
		_ = os.Remove(prevReportFilePath)
		_ = os.Rename(reportFilePath, prevReportFilePath)
	}
}

func SaveReport(report string, logsFolder string, dryRun bool) {
	if !dryRun {
		f, err := os.OpenFile(logsFolder+"simple_backup-go-report.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("Error when opening report file:", err)
		}
		if _, err := f.WriteString("\n" + report); err != nil {
			log.Fatal("Error when writing report file:", err)
		}
		if err := f.Close(); err != nil {
			log.Fatal("Error when closing report file:", err)
		}
	}
}

func validate7zOutput(output string) bool {
	return strings.Contains(strings.ToUpper(output), "EVERYTHING IS OK")
}

// Parallelize runs given functions in parallel.
func Parallelize(functions ...func()) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(functions))

	defer waitGroup.Wait()

	for _, function := range functions {
		go func(copy func()) {
			defer waitGroup.Done()
			copy()
		}(function)
	}
}

// BackupTaskConfigs backup tasks. Generated with https://mholt.github.io/json-to-go/
type BackupTaskConfigs []struct {
	TaskName            string   `json:"task-name"`
	Source              string   `json:"source"`
	ProtectWithPassword string   `json:"protect-with-password,omitempty"`
	Excludes            []string `json:"excludes,omitempty"`
}

// LoadBackupTaskConfigs loads given config file then returns configured backup tasks.
func LoadBackupTaskConfigs(configFile string) BackupTaskConfigs {
	content, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error when opening file:", err)
	}
	// Now let's unmarshall the data into `config`
	var config BackupTaskConfigs
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal():", err)
	}
	return config
}

func IsElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
