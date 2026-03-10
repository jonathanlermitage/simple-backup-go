<h1 align="center">
    <a href="https://plugins.jetbrains.com/plugin/11058-extra-icons">
      <img src="./media/go.svg" width="84" height="84" alt="logo"/>
    </a><br/>
    Simple Backup Go
</h1>

A simple command-line backup tool written in Go, for my personal needs. I used to write CMD files that compress many folders with 7zip (with sub-folders filters, password protection...), and I wanted to move all this stuff to a single JSON file.    
This program takes some command-line parameters and a JSON config file, and generates archive files in desired folder.

Windows: tested on Windows 10 x64 with 7zip 22.01, 24.09, 25.01, and 26.00. [7zip](https://www.7-zip.org/) should be in PATH.

<!--Linux: tested on Arch Linux with [7zip-25.01-1](https://archlinux.org/packages/extra/x86_64/7zip/). Other Linux distributions may work too, like Debian or Fedora.-->

## Build

Simply run `go build` or `make build` and see generated executable.

## Usage

The program takes command-line parameters:

| Command            | Description                                                                        | Default |                          Required                          |
|--------------------|------------------------------------------------------------------------------------|:-------:|:----------------------------------------------------------:|
| targetFolder       | where the backups are located                                                      |         |                            yes                             |
| workFolder            | temporary work directory for archive generation                                    |         |                             no                             |
| configFile         | the location of the JSON config file                                               |         |                            yes                             |
| logsFolder         | where the logs are saved (will save no logs if not specified)                      |         |                             no                             |
| restartOneDrive    | stop MS OneDrive before backing up, then start it once the backup is complete      | `false` |                             no                             |
| password           | password used to protect backups, if password required by JSON config              |         | if `protect-with-password` is set to `true` in JSON config |
| taskNamesToExecute | backup tasks filter: use it to execute only given backup tasks (see example below) |         |                             no                             |

A JSON config file describes one or many backup tasks. See the sample [backup.json file](./sample/backup.json). It can be either an array of tasks (legacy format) or an object containing a list of tasks and optional default values.

### Root-level fields

| Field             | Description                                                           | Default | Required |
|-------------------|-----------------------------------------------------------------------|:-------:|:--------:|
| compression-ratio | default compression level for all tasks (7zip `-mx` parameter)         |  `-mx3` |    no    |
| tasks             | a list of backup tasks (required if the JSON is an object)           |         |   yes    |

### Task-level fields

Every backup task is described by some fields:

| Field                 | Description                                                                       | Default | Required |
|-----------------------|-----------------------------------------------------------------------------------|:-------:|:--------:|
| task-name             | a name for the backup task                                                        |         |   yes    |
| source                | the folder to back up                                                             |         |   yes    |
| protect-with-password | protect the generated archive file with the password provided by the command line | `false` |    no    |
| compression-ratio    | set the compression level (7zip `-mx` parameter)                                   |  `-mx3` |    no    |
| excludes              | a list of sub-folders to exclude from the backup                                  |  `[]`   |    no    |

Examples: 

```shell
# execute the backup tasks defined by backup-config.json
simple-backup-go.exe -password=foo -targetFolder=D:\data\my_backups -configFile=C:\Users\jonathan\backup-config.json

# execute only the "Foobar2000 playlists" and "Stardew Valley profile and saves" backup tasks
simple-backup-go.exe -password=foo -targetFolder=D:\data\my_backups -configFile=C:\Users\jonathan\backup-config.json -taskNamesToExecute="Foobar2000 playlists,Stardew Valley profile and saves"
```

It will generate archive files in the target folder. Archive files look like this: `the_source_folder_path YYYYMMDD hhmmss.7z` (Windows) or `.zip` (Linux).  
Example: `C_Projects 20220819 201536.7z`.  
Compression is set to _Fast_ (`-mx3` 7zip parameter) by default, but it can be overridden by the root-level `compression-ratio` field, or the task-level `compression-ratio` field. It should also compress files open for writing (`-ssw` 7zip parameter).

Tip: you may want to start the program in a new tab on Windows Terminal (`wt`) instead of the classical CMD console (which doesn't render emojis correctly). Please use a command like this: `wt -w 0 nt C:\theAbsolutePathOf\simple-backup-go.exe ...arguments...`, otherwise you may see an error when Windows Terminal is already opened. See the [Windows Terminal documentation](https://docs.microsoft.com/en-us/windows/terminal/command-line-arguments?tabs=windows#open-a-new-profile-instance).

## License

MIT License. In other words, you can do what you want: this project is entirely OpenSource, Free and Gratis.  
