package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func addProxy(domain string, endpoint string, uriLocation string, headers map[string]string) bool {
	if !siteExists(domain) {
		fmt.Println("Aborted: Site does not exist: " + domain)
		return false
	}
	var site SiteInfo
	newLocation := ReverseProxyLocation{
		URLLocation: uriLocation,
		Endpoint:    endpoint,
		Headers:     headers,
	}
	site = getSiteByDomain(domain)
	site.ProxyLocations = append(site.ProxyLocations, newLocation)
	updateSite(site)
	return true
}

func addStaticSite(domain string, rootPath string, uriLocation string) bool {
	if !siteExists(domain) {
		fmt.Println("Aborted: Site does not exist: " + domain)
		return false
	}
	if !dirExists(rootPath) {
		ensureDirExists(rootPath)
		content := "<h1>It's working! (" + domain + ")</h1>"
		try(writeFile(rootPath+"/index.html", []byte(content)))
	}
	var site SiteInfo
	// count the number of dots in the domain name
	newLocation := StaticLocation{
		URLLocation: uriLocation,
		RootPath:    rootPath,
	}
	site = getSiteByDomain(domain)
	site.StaticLocations = append(site.StaticLocations, newLocation)
	updateSite(site)
	return true
}

func isSubDomain(domain string) bool {
	dots := strings.Count(domain, ".")
	return dots > 1
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

func createSite(domain string, rootPath string) {
	if !siteExists(domain) {
		subDomain := isSubDomain(domain)
		site := initSite(domain, subDomain)
		site.RootPath = rootPath
		updateSite(site)
	} else {
		fmt.Println("Site already exists: " + domain)
	}
}
