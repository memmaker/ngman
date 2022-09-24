package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
		fmt.Println("Warning: " + err.Error())
	}
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ensureDirExists(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		fmt.Println("Directory '" + dir + "' does not exist, creating it")
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

func readLines(filename string) []string {
	scanner := bufio.NewScanner(must(os.Open(filename)))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func writeFile(filename string, content []byte) error {
	return os.WriteFile(filename, content, 0644)
}
