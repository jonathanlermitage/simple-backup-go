package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const OneK = 1024
const OneM = 1024 * 1024
const OneG = 1024 * 1024 * 1024

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
func Compress(folderToCompress string, excludes []string, targetFolder string, archiveName string, password string, protectWithPassword bool, workFolder string, compressionRatio string) string {
	var archiveExtension = ".7z"
	var archiveFormat = "-t7z"
	var p7zipCommand = "7z"

	if runtime.GOOS == "windows" {
		p7zipCommand = "7z.exe"
	}

	args := []string{"a", archiveFormat}

	if protectWithPassword && len(password) > 0 {
		args = append(args, "-p"+password)
	}

	folderToCompress = strings.Replace(folderToCompress, "\\", "/", -1)
	if strings.HasSuffix(folderToCompress, "/") {
		folderToCompress = folderToCompress[0 : len(folderToCompress)-1]
	}

	var targetFolderForCurrentMonth = strings.Replace(targetFolder, "\\", "/", -1)
	if !(strings.HasSuffix(targetFolderForCurrentMonth, "/")) {
		targetFolderForCurrentMonth += "/"
	}
	targetFolderForCurrentMonth += GetCurrentDate() + "/"

	var finalFileName = archiveName + " " + GetCurrentDateTime() + archiveExtension
	var archivePath = targetFolderForCurrentMonth + finalFileName

	var tempArchivePath = ""
	if len(workFolder) > 0 {
		workFolder = strings.Replace(workFolder, "\\", "/", -1)
		if !(strings.HasSuffix(workFolder, "/")) {
			workFolder += "/"
		}
		tempArchivePath = workFolder + finalFileName
	}

	if !FolderExists(targetFolderForCurrentMonth) {
		mkdirErr := os.MkdirAll(targetFolderForCurrentMonth, 0755)
		if mkdirErr != nil {
			log.Fatal(mkdirErr)
		}
	}
	if len(workFolder) > 0 {
		if !FolderExists(workFolder) {
			mkdirErr := os.MkdirAll(workFolder, 0755)
			if mkdirErr != nil {
				log.Fatal(mkdirErr)
			}
		}
	}

	if FileExists(archivePath) {
		_ = os.Remove(archivePath)
	}
	if len(tempArchivePath) > 0 && FileExists(tempArchivePath) {
		_ = os.Remove(tempArchivePath)
	}

	var pathToUseFor7z = archivePath
	if len(tempArchivePath) > 0 {
		pathToUseFor7z = tempArchivePath
	}

	compressionRatioToUse := determineCompressionRatio(compressionRatio)

	args = append(args, pathToUseFor7z, folderToCompress,
		"-ssw",                /* compress files open for writing */
		compressionRatioToUse, /* set level of compression */
		"-bd"                  /* disable progress indicator */)

	var firstFolder = folderToCompress[strings.LastIndex(folderToCompress, "/")+1:]
	for i := 0; i < len(excludes); i++ {
		args = append(args, "-xr!"+firstFolder+"/"+excludes[i])
	}

	out, err := exec.Command(p7zipCommand, args...).Output()

	if err != nil {
		fmt.Println(err)
	}

	var compressionResult = string(out[:])

	if validate7zOutput(compressionResult) {
		if len(tempArchivePath) > 0 {
			if FileExists(tempArchivePath) {
				fmt.Println("Moving archive from " + tempArchivePath + " to " + archivePath)
				moveErr := os.Rename(tempArchivePath, archivePath)
				if moveErr != nil {
					fmt.Printf("Error moving archive: %v\n", moveErr)
					// If rename fails (e.g. cross-device), we might need another way,
					// but os.Rename is usually fine within the same filesystem or between some.
					// However, the requirement is to move it.
				}
			}
		}

		if FileExists(archivePath) {
			fmt.Println("Generated archive file " + archivePath + " (approx. " + GetFileSize(archivePath) + ")")
			PrintBackupSizeEvolutionOverTime(targetFolderForCurrentMonth, archiveName, archivePath)
		}
	} else {
		if len(tempArchivePath) > 0 && FileExists(tempArchivePath) {
			_ = os.Remove(tempArchivePath)
			fmt.Println("Compression failed, deleted bad temporary archive file " + tempArchivePath)
		} else if FileExists(archivePath) {
			_ = os.Remove(archivePath)
			fmt.Println("Compression failed, deleted bad archive file " + archivePath)
		}
	}

	var readableCompressionString = p7zipCommand
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-p") {
			readableCompressionString += " -p***"
		} else {
			readableCompressionString += " " + args[i]
		}
	}

	return readableCompressionString + "\n" + compressionResult
}

func PrintBackupSizeEvolutionOverTime(archivesBaseFolder string, archivesBaseName string, lastFilePath string) {
	//
}

func GetFileSize(filePath string) string {
	fi, err := os.Stat(filePath)
	if err != nil {
		log.Fatal("Error when evaluating file's size:", err)
	}
	var fiSize = fi.Size()
	var sizeUnit = "B"
	if fiSize > OneG {
		fiSize = fiSize / OneG
		sizeUnit = "GB"
	} else if fiSize > OneM {
		fiSize = fiSize / OneM
		sizeUnit = "MB"
	} else if fiSize > OneK {
		fiSize = fiSize / OneK
		sizeUnit = "KB"
	}
	return fmt.Sprintf("%d%s", fiSize, sizeUnit)
}

func RotateLogs(logsFolder string) {
	if len(logsFolder) > 0 {
		var prevReportFilePath = logsFolder + "simple-backup-go-logs_prev.txt"
		var reportFilePath = logsFolder + "simple-backup-go-logs.txt"
		fmt.Println("🧾 Report rotation: move " + reportFilePath + " to " + prevReportFilePath)
		_ = os.Remove(prevReportFilePath)
		_ = os.Rename(reportFilePath, prevReportFilePath)
	}
}

func SaveLogs(report string, logsFolder string) {
	if len(logsFolder) > 0 {
		f, err := os.OpenFile(logsFolder+"simple-backup-go-logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("Error when opening log file:", err)
		}
		if _, err := f.WriteString("\n##########\n" + report); err != nil {
			log.Fatal("Error when writing log file:", err)
		}
		if err := f.Close(); err != nil {
			log.Fatal("Error when closing log file:", err)
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

type BackupTaskConfig struct {
	TaskName            string   `json:"task-name"`
	Source              string   `json:"source"`
	ProtectWithPassword string   `json:"protect-with-password,omitempty"`
	CompressionRatio    string   `json:"compression-ratio,omitempty"`
	Excludes            []string `json:"excludes,omitempty"`
}

type BackupConfigs struct {
	CompressionRatio string             `json:"compression-ratio,omitempty"`
	Tasks            []BackupTaskConfig `json:"tasks"`
}

// LoadBackupTaskConfigs loads given config file then returns configured backup tasks.
func LoadBackupTaskConfigs(configFile string) BackupConfigs {
	content, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error when opening file:", err)
	}
	// Now let's unmarshall the data into `config`
	var config BackupConfigs
	err = json.Unmarshal(content, &config)
	if err == nil && len(config.Tasks) > 0 {
		return config
	}

	// Try old format
	var tasks []BackupTaskConfig
	err = json.Unmarshal(content, &tasks)
	if err == nil {
		return BackupConfigs{Tasks: tasks}
	}

	log.Fatal("Error during Unmarshal():", err)
	return BackupConfigs{}
}

func IsElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func determineCompressionRatio(ratio string) string {
	var compressionRatioToUse = "-mx3"
	if len(ratio) > 0 {
		if strings.HasPrefix(ratio, "-mx") {
			compressionRatioToUse = ratio
		} else {
			compressionRatioToUse = "-mx" + ratio
		}
	}
	return compressionRatioToUse
}
