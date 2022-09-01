<h1 align="center">
    <a href="https://plugins.jetbrains.com/plugin/11058-extra-icons">
      <img src="./media/go.svg" width="84" height="84" alt="logo"/>
    </a><br/>
    Simple Backup Go
</h1>

A simple command-line backup tool written in Go, for my personal needs. I used to write CMD files that compress many folders with 7zip (with sub-folders filters, password protection...), and I wanted to move all this stuff to a single JSON file.    
This program takes some command-line parameters and a JSON config file, and generates archive files in desired folder.

[7zip](https://www.7-zip.org/) (7z.exe) should be in PATH. Tested on Windows 10 x64 with 7zip 22.01, and compiled with Go 1.19.

## Build

Simply run `go build` or `make build` and see generated executable.

## Usage

The program takes command-line parameters:

| Command            | Description                                                                        | Default |                          Required                          |
|--------------------|------------------------------------------------------------------------------------|:-------:|:----------------------------------------------------------:|
| targetFolder       | where the backups are located                                                      |         |                            yes                             |
| configFile         | the location of the JSON config file                                               |         |                            yes                             |
| logsFolder         | where the logs are saved (will save no logs if not specified)                      |         |                             no                             |
| dryRun             | don't run the backup tasks, and show what the 7zip commands would be instead       | `false` |                             no                             |
| password           | password used to protect backups, if password required by JSON config              |         | if `protect-with-password` is set to `true` in JSON config |
| taskNamesToExecute | backup tasks filter: use it to execute only given backup tasks (see example below) |         |                             no                             |

A JSON config file describes one or many backup tasks. See the sample [backup.json file](./sample/backup.json). Every backup task is described by some fields:

| Field                 | Description                                                                       | Default | Required |
|-----------------------|-----------------------------------------------------------------------------------|:-------:|:--------:|
| task-name             | a name for the backup task                                                        |         |   yes    |
| source                | the folder to back up                                                             |         |   yes    |
| protect-with-password | protect the generated archive file with the password provided by the command line | `false` |    no    |
| excludes              | a list of sub-folders to exclude from the backup                                  |  `[]`   |    no    |

Examples: 

```shell
# execute the backup tasks defined by backup-config.json
simple-backup-go.exe -password=foo -targetFolder=D:\data\my_backups -configFile=C:\Users\jonathan\backup-config.json

# execute only the "Foobar2000 playlists" and "Stardew Valley profile and saves" backup tasks
simple-backup-go.exe -password=foo -targetFolder=D:\data\my_backups -configFile=C:\Users\jonathan\backup-config.json -taskNamesToExecute="Foobar2000 playlists,Stardew Valley profile and saves"
```

It will generate 7zip archive files in the target folder. Archive files look like this: `the_source_folder_path YYYYMMDD hhmmss.7z`, pex example `C_Projects 20220819 201536.7z`.  
Compression is set to _Fast_ (`-mx3` 7zip parameter) and it should compress files open for writing (`-ssw` 7zip parameter).

## To-Do

* run backup tasks in parallel (make it configurable?)
* don't stop on first error (make it configurable?)
* move Go files to a folder like `src/main` and update the build command (I'm very new to Go programming...😅)

## License

MIT License. In other words, you can do what you want: this project is entirely OpenSource, Free and Gratis.  
