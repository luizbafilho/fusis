package net

import (
	"errors"
	"io/ioutil"
	"strings"
)

const (
	sysctlDir = "/proc/sys/"
)

var invalidKeyError = errors.New("could not find the given key")

func GetSysctl(name string) (string, error) {
	path := sysctlDir + strings.Replace(name, ".", "/", -1)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", invalidKeyError
	}
	return strings.TrimSpace(string(data)), nil
}

func SetSysctl(name string, value string) error {
	path := sysctlDir + strings.Replace(name, ".", "/", -1)
	return ioutil.WriteFile(path, []byte(value), 0644)
}
