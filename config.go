package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"gopkg.in/yaml.v2"
)

const (
	BACKUP_EXT     = ".bak"
	EXT_YAML       = ".yaml"
	EXT_JSON       = ".json"
	USE_CONFIG_YML = true
	DATA_EXT       = ".data"
	LOG_EXT        = ".log"
)

var (
	ErrTokenUndefined  = errors.New("token is empty")
	ErrMasterUndefined = errors.New("master id is not set")
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
	pfx := AppBaseFileName() + "-config"
	if USE_CONFIG_YML {
		return pfx + EXT_YAML
	} else {
		return pfx + EXT_JSON
	}
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

func LoadConfig() error {
	cfname := GetConfigName()
	f, err := os.Open(cfname)
	if err != nil {
		return err
	}
	defer f.Close()
	if USE_CONFIG_YML {
		err = yaml.NewDecoder(f).Decode(&config)
	} else {
		err = json.NewDecoder(f).Decode(&config)
	}
	if err != nil {
		return err
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
	return nil
}

func MakeBackup(fname string) {
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		return
	}
	// make backup
	fname_backup := fname + ".bak"
	if _, err := os.Stat(fname_backup); !os.IsNotExist(err) {
		err = os.Remove(fname_backup)
		if err != nil {
			log.Fatalf("Remove(%s) failed: %s", fname_backup, err)
		}
	}
	if err := os.Rename(fname, fname_backup); err != nil {
		log.Fatalf("Rename(%s, %s) failed: %s", fname, fname_backup, err)
	}
}

func SaveConfig() error {
	cfname := GetConfigName()
	MakeBackup(cfname)
	f, err := os.Create(cfname)
	defer f.Close()
	if err != nil {
		return err
	}
	if USE_CONFIG_YML {
		err = yaml.NewEncoder(f).Encode(config)
	} else {
		err = json.NewEncoder(f).Encode(config)
	}
	return err
}

func ValidateConfig() error {
	if config.Token == "" {
		return ErrTokenUndefined
	}
	if config.Master == 0 {
		return ErrMasterUndefined
	}
	return nil
}
