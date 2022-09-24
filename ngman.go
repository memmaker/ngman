package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"text/template"
)

var config GlobalConfig

var rootTemplate *template.Template

func main() {
	loadConfig()
	if config.WebRootPath == "" {
		printCheckConfig()
		log.Fatal("Web root path is not set.")
	}
	if config.SiteStorageDirectory == "" {
		printCheckConfig()
		log.Fatal("Site storage directory is not set.")
	}
	if config.NginxSiteConfigDirectory == "" {
		printCheckConfig()
		log.Fatal("Nginx site config directory is not set.")
	}
	if !fileExists(config.TemplateFile) {
		printCheckConfig()
		log.Fatal("Template file '" + config.TemplateFile + "' does not exist")
	}
	ensureDirExists(config.WebRootPath)
	ensureDirExists(config.SiteStorageDirectory)
	ensureDirExists(config.NginxSiteConfigDirectory)

	if !dirExists(config.WebRootPath) || !dirExists(config.SiteStorageDirectory) || !dirExists(config.NginxSiteConfigDirectory) {
		printCheckConfig()
		log.Fatal("Could not ensure that all directories exist, check access rights.")
	}

	loadTemplates()

	// get command line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return
	}
	if args[0] == "add-site" && len(args) > 1 {
		var rootPath string
		if len(args) == 3 {
			rootPath = args[2]
		} else {
			rootPath = path.Join(config.WebRootPath, args[1])
			fmt.Println("No WebRoot specified, using '" + rootPath + "'")
		}
		createSite(args[1], rootPath)
		return
	}
	if args[0] == "add-static" && len(args) == 4 {
		addStaticSite(args[1], args[2], args[3])
		tryPostRunCommand()
		return
	}
	if args[0] == "add-proxy" && len(args) > 2 {
		var uriLocation string
		if len(args) == 4 {
			uriLocation = args[3]
		} else {
			uriLocation = "/"
			fmt.Println("No URI location specified, using '" + uriLocation + "'")
		}
		addProxy(args[1], args[2], uriLocation, nil)
		tryPostRunCommand()
		return
	}
	if args[0] == "delete" && len(args) == 2 {
		deleteSite(args[1])
		tryPostRunCommand()
		return
	}
	if args[0] == "edit" && len(args) == 2 {
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
	printUsage()
}

func printCheckConfig() {
	configFilename := path.Join(getConfigDir(), "config.toml")
	fmt.Println("Please check the config file at " + configFilename)
}

func printUsage() {
	fmt.Println("Usage: ngman list")
	fmt.Println("Usage: ngman add-site <domain> [<webroot>]")
	fmt.Println("Usage: ngman add-static <domain> <webroot> <uri-location>")
	fmt.Println("Usage: ngman add-proxy <domain> <endpoint> [<uri-location>]")
	fmt.Println("Usage: ngman edit <domain>")
	fmt.Println("Usage: ngman delete <domain>")
	fmt.Println("Usage: ngman write-all")
}
