package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func addProxy(domain string, endpoint string, uriLocation string, headers map[string]string) bool {
	ensureSiteExists(domain)

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
	ensureSiteExists(domain)

	if !dirExists(rootPath) {
		ensureDirExists(rootPath)
		content := "<h1>It's working! (" + domain + ")</h1>"
		try(writeFile(path.Join(rootPath, "index.html"), []byte(content)))
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

func ensureSiteExists(domain string) {
	if !siteExists(domain) {
		rootPath := path.Join(config.WebRootPath, domain)
		fmt.Println("No WebRoot specified, using '" + rootPath + "'")
		createSite(domain, rootPath)
	}
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
	cmd := exec.Command(editor, path.Join(config.SiteStorageDirectory, domain+".toml"))
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
	try(os.Remove(path.Join(config.SiteStorageDirectory, domain+".toml")))
	try(os.Remove(path.Join(config.NginxSiteConfigDirectory, domain+".conf")))
	path.Join()
}

func createSite(domain string, rootPath string) {
	if !siteExists(domain) {
		initSite(domain, rootPath)
	} else {
		fmt.Println("Site already exists: " + domain)
	}
}
