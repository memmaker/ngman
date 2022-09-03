package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func addProxy(domain string, endpoint string, uriLocation string, headers map[string]string) {
	var site SiteInfo
	newLocation := ReverseProxyLocation{
		URLLocation: uriLocation,
		Endpoint:    endpoint,
		Headers:     headers,
	}
	if !siteExists(domain) {
		site = initSite(domain, true)
		site.ProxyLocations = []ReverseProxyLocation{newLocation}
	} else {
		site = getSiteByDomain(domain)
		site.ProxyLocations = append(site.ProxyLocations, newLocation)
	}
	updateSite(site)
}

func addStaticSite(domain string, rootPath string, uriLocation string) {
	if !dirExists(rootPath) {
		ensureDirExists(rootPath)
		content := "<h1>It's working! (" + domain + ")</h1>"
		try(writeFile(rootPath+"/index.html", []byte(content)))
	}
	var site SiteInfo
	// count the number of dots in the domain name
	dots := strings.Count(domain, ".")
	newLocation := StaticLocation{
		URLLocation: uriLocation,
		RootPath:    rootPath,
	}
	if !siteExists(domain) {
		site = initSite(domain, dots > 1)
		site.StaticLocations = []StaticLocation{newLocation}
	} else {
		site = getSiteByDomain(domain)
		site.StaticLocations = append(site.StaticLocations, newLocation)
	}
	updateSite(site)
}

func editSite(domain string) {
	// call the editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	cmd := exec.Command(editor, config.SiteStorageDirectory+"/"+domain+".json")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	site := getSiteByDomain(domain)
	writeNginxConfig(site)
	fmt.Println("Site updated: " + domain)
}

func listAll() {
	sites := getAllSites()
	for _, site := range sites {
		if site.RootPath != "" {
			fmt.Println(site.Domain + " -> " + site.RootPath)
		} else {
			fmt.Println(site.Domain)
		}
		// print all static locations
		for _, location := range site.StaticLocations {
			fmt.Println("  " + location.URLLocation + " -> " + location.RootPath)
		}
		// print all proxy locations
		for _, location := range site.ProxyLocations {
			fmt.Println("  " + location.URLLocation + " -> " + location.Endpoint)
		}
	}
}

func writeAll() {
	sites := getAllSites()
	for _, site := range sites {
		writeNginxConfig(site)
	}
}

func deleteSite(domain string) {
	try(os.Remove(config.SiteStorageDirectory + "/" + domain + ".json"))
	try(os.Remove(config.NginxSiteConfigDirectory + "/" + domain + ".conf"))
}
