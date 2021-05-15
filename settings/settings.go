package settings

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type conf struct {
	TargetFileSize       string         `yaml:"targetFileSize"`
	Workers              int          	`yaml:"workers"`
	NewFilename          string         `yaml:"newFilename"`
	TargetFolder         string         `yaml:"targetFolder"`
	TryRotateByFace      float32        `yaml:"tryRotateByFace"`
}

var Config conf

// FileExists checks a file's existence
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func confLoader(c interface{}, configFile string) interface{} {
	if FileExists(configFile) == true {
		yamlFile, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, c)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	} else {
		log.Printf("settings.yaml not found (Press Enter to exit)")
		fmt.Scanln()
		os.Exit(1)
	}

	return c
}

func (c *conf) GetConf(configFile string) *conf {
	return confLoader(c, configFile).(*conf)
}
