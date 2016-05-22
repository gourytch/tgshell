package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func ExeName() string {
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	return exe
}

func AppDir() string {
	dir, err := filepath.Abs(filepath.Dir(ExeName()))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func AppBaseFileName() string {
	r := regexp.MustCompile("^(.*?)(?:\\.exe|\\.EXE|)$")
	return r.FindStringSubmatch(ExeName())[1]
}

func GetConfigName() string {
	return AppBaseFileName() + ".config"
}

func CheckDatadir() {
	if config.Data_Dir != "" {
		if err := os.MkdirAll(config.Data_Dir, 0755); err != nil {
			log.Fatalf("mkDirAll(%s) got error: %s", config.Data_Dir, err)
		}
	}
}
func DumpConfig() {
	log.Printf("current config:\n%s", ppj(config))

}

func LoadConfig() {
	fname := GetConfigName()
	log.Printf("Load config from '%s' ...", fname)
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("... loaded")
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
	data, err := json.MarshalIndent(config, "", "\t")
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
	log.Printf("Save config to '%s' ...", fname)
	err = ioutil.WriteFile(fname, data, 0600)
	if err != nil {
		log.Printf("... error:", err)
	} else {
		log.Printf("... saved")
	}
}
