package main

import (
	"log"
	"os"
	"os/user"
	"strings"
)

func removeRedundantEmptyLines(content string) string {
	return strings.Replace(content, "\n\n", "\n", -1)
}

func must[V any](value V, err error) V {
	if err != nil {
		log.Fatal(err)
	}
	return value
}
func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ensureDirExists(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		try(os.Mkdir(dir, 0755))
	}
}

func readFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}

func writeFile(filename string, content []byte) error {
	return os.WriteFile(filename, content, 0644)
}
