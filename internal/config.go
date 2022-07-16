//     go-weather-reporter: pull from weather service, push to database
//     Copyright (C) 2022 Josh Simonot
//
//     This program is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     This program is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with this program.  If not, see <https://www.gnu.org/licenses/>.

package internal

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type ConfigParser struct {
	logr *log.Logger
}

type Opts struct {
	ConfigDir string
	Once      bool
}

type Config []ServiceConfig

type ServiceConfig struct {
	Name         string
	ConfPath     string
	Source       map[string]interface{}
	Destinations []map[string]interface{}
}

func NewConfigParser(logr *log.Logger) *ConfigParser {
	return &ConfigParser{logr: logr}
}

func (c *ConfigParser) ParseConfigFiles(dir string) (Config, error) {
	var conf Config

	dirents, err := os.ReadDir(dir)
	if err != nil {
		return conf, err
	}

	for _, dirent := range dirents {
		if !dirent.IsDir() {

			path := filepath.Join(dir, dirent.Name())
			switch filepath.Ext(dirent.Name()) {

			case ".yaml":
				c.logr.Println(dirent.Name())
				content, err := os.ReadFile(path)
				if err != nil {
					return conf, err
				}
				content = EnvVarSubstitution(content)
				c, err := ParseYamlConfig(content)
				if err != nil {
					return conf, err
				}
				for _, service := range c {
					service.ConfPath = path
				}
				conf = append(conf, c...)

			default:
				c.logr.Println("info: skipping file", dirent.Name())
			}
		}
	}
	return conf, nil
}

func ParseYamlConfig(src []byte) (Config, error) {
	var conf Config
	err := yaml.UnmarshalStrict([]byte(src), &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

func EnvVarSubstitution(in []byte) []byte {
	rexp, err := regexp.Compile(`\${\w*}`)
	if err != nil {
		log.Fatalln(err)
	}
	lines := strings.Split(string(in), "\n")
	for i, line := range lines {

		// # ignored empty and commented lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		matches := rexp.FindAllString(line, -1)
		for _, match := range matches {

			// special value, must not be replaced
			if match == "${field}" {
				continue
			}
			envKey := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
			sub := os.Getenv(envKey)
			if sub == "" {
				log.Fatalln("error: missing environment variable: ", match)
			}
			line = strings.ReplaceAll(line, match, sub)
		}
		lines[i] = line
	}
	return []byte(strings.Join(lines, "\n"))
}
