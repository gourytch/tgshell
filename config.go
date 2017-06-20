package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
)

const (
	BACKUP_EXT = ".bak"
	CONFIG_EXT = ".config"
	DATA_EXT   = ".data"
	LOG_EXT    = ".log"
)

func RealPath(path string) string {
	r, _ := filepath.Abs(filepath.FromSlash(path))
	return r
}

func TrimLastExt(fname string) string {
	name := filepath.Base(fname)
	ext := filepath.Ext(name)
	lfname := len(fname)
	lname := len(name)
	lext := len(ext)
	if lext == 0 || lext == lname {
		return fname
	}
	return fname[0 : lfname-lext]
}

func ExeName() string {
	return RealPath(os.Args[0])
}

func AppDir() string {
	return filepath.Dir(RealPath(ExeName()))
}

func AppBaseFileName() string {
	return TrimLastExt(ExeName())
}

func GetConfigName() string {
	return AppBaseFileName() + CONFIG_EXT
}

func GetLogName() string {
	return AppBaseFileName() + LOG_EXT
}

func SetupLogger() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   GetLogName(),
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28, //days
	})
}

func CheckDatadir() {
	if config.Data_Dir != "" {
		if err := os.MkdirAll(config.Data_Dir, 0755); err != nil {
			log.Fatalf("mkDirAll(%s) got error: %s", config.Data_Dir, err)
		}
	}
}

func LoadConfig() {
	data, err := ioutil.ReadFile(GetConfigName())
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	config.Host, err = os.Hostname()
	if err != nil {
		log.Printf("os.Hostname() error: %s", err)
		config.Host = "unknown"
	}
	if config.Data_Dir == "" {
		config.Data_Dir = AppDir() + "/data"
		CheckDatadir()
		SaveConfig()
	}
}

func SaveConfig() {
	data, err := json.MarshalIndent(config, "", "  ") // json.Marshal(config)

	if err != nil {
		log.Fatalf("Marshal error: %s", err)
	}
	fname := GetConfigName()
	if _, err = os.Stat(fname); !os.IsNotExist(err) {
		// make backup
		fname_backup := fname + ".bak"
		if _, err := os.Stat(fname_backup); !os.IsNotExist(err) {
			err = os.Remove(fname_backup)
			if err != nil {
				log.Fatalf("Remove(%s) failed: %s", fname_backup, err)
			}
		}
		err = os.Rename(fname, fname_backup)
		if err != nil {
			log.Fatalf("Rename(%s, %s) failed: %s", fname, fname_backup, err)
		}
	}
	err = ioutil.WriteFile(fname, data, 0600)
}
