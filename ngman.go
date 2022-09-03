package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

var config GlobalConfig

var rootTemplate *template.Template

func main() {
	if !isRoot() {
		log.Fatal("This program must be run as root")
	}
	loadConfig()
	ensureDirExists(config.SiteStorageDirectory)
	ensureDirExists(config.NginxSiteConfigDirectory)
	loadTemplates()

	// get command line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: ngman list")
		fmt.Println("Usage: ngman edit <domain>")
		fmt.Println("Usage: ngman add-static <domain> <root-path> <uri-location>")
		fmt.Println("Usage: ngman add-proxy <domain> <endpoint> <uri-location>")
		fmt.Println("Usage: ngman delete <domain>")
		fmt.Println("Usage: ngman write-all")
		return
	}
	if args[0] == "add-static" {
		addStaticSite(args[1], args[2], args[3])
		tryPostRunCommand()
		return
	}
	if args[0] == "add-proxy" {
		addProxy(args[1], args[2], args[3], nil)
		tryPostRunCommand()
		return
	}
	if args[0] == "delete" {
		deleteSite(args[1])
		tryPostRunCommand()
		return
	}
	if args[0] == "edit" {
		editSite(args[1])
		tryPostRunCommand()
		return
	}
	if args[0] == "list" {
		listAll()
		return
	}
	if args[0] == "write-all" {
		writeAll()
		tryPostRunCommand()
		return
	}
	fmt.Println("Unknown command: " + args[0])
}
