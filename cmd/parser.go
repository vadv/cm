package main

// по идее тут надо использовать нормальный язычок разметки
// но все что я пробовал обладают недостатками :)
// json - для человека слишком сложен
// yaml - не имеет примитивов
// ini - не умел val не в strings

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	sectionRegex   = regexp.MustCompile(`^\[(.*)\]$`)
	assignRegex    = regexp.MustCompile(`^([^=]+)=(.*)$`)
	boolTrueRegex  = regexp.MustCompile(`^(true|on)$`)
	boolFalseRegex = regexp.MustCompile(`^(false|on)$`)
	floatRegex     = regexp.MustCompile(`^[-+]?[0-9]*(\.|\,)[0-9]+([eE][-+]?[0-9]+)?$`)
	intRegex       = regexp.MustCompile(`^[-+]?[0-9]*$`)
)

type ConfigFile map[string]ConfigSection
type ConfigSection map[string]interface{}

func parseFile(filename string) (ConfigFile, error) {

	in, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open config: %s", err.Error())
	}
	defer in.Close()
	bufin := bufio.NewReader(in)

	file := make(map[string]ConfigSection)

	section := ""
	lineNum := 0

	for done := false; !done; {
		var line string
		if line, err = bufin.ReadString('\n'); err != nil {
			if err == io.EOF {
				done = true
			} else {
				return file, nil
			}
		}
		lineNum++
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			// пустые
			continue
		}
		if line[0] == ';' || line[0] == '#' {
			// комментарии
			continue
		}

		if groups := assignRegex.FindStringSubmatch(line); groups != nil {
			key, val := groups[1], groups[2]
			key, val = strings.TrimSpace(key), strings.TrimSpace(val)

			if val[0] == '"' {
				// string
				if val[len(val)-1] != '"' {
					return nil, fmt.Errorf("invalid INI syntax on line %d: %s, unexcepted '\"'", lineNum, line)
				}
				val = strings.Trim(val, `"`)
				file[section][key] = val
			} else {
				// not string
				switch {
				//bool
				case boolFalseRegex.MatchString(strings.ToLower(val)):
					file[section][key] = false
				//bool
				case boolTrueRegex.MatchString(strings.ToLower(val)):
					file[section][key] = true
				//float
				case floatRegex.MatchString(val):
					if res, err := strconv.ParseFloat(val, 64); err == nil {
						file[section][key] = res
					} else {
						return nil, fmt.Errorf("invalid INI syntax on line %d: %s, can't parse float value", lineNum, line)
					}
				//int
				case intRegex.MatchString(val):
					if res, err := strconv.ParseInt(val, 10, 64); err == nil {
						file[section][key] = res
					} else {
						return nil, fmt.Errorf("invalid INI syntax on line %d: %s, can't parse integer value", lineNum, line)
					}
				default:
					return nil, fmt.Errorf("invalid INI syntax on line %d: %s\nUnknown type of value (string?): %s", lineNum, line, val)
				}
			}

		} else if groups := sectionRegex.FindStringSubmatch(line); groups != nil {
			name := strings.TrimSpace(groups[1])
			section = name
			if _, ok := file[section]; !ok {
				file[section] = make(map[string]interface{})
			}
		} else {
			return nil, fmt.Errorf("invalid INI syntax on line %d: %s", lineNum, line)
		}
	}
	return file, nil
}
