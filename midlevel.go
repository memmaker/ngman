package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

func getAllSites() []SiteInfo {
	var sites []SiteInfo
	files, err := os.ReadDir(config.SiteStorageDirectory)
	if err != nil {
		return sites
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		siteFilename := file.Name()
		var domainOfSite string
		if strings.HasSuffix(siteFilename, ".toml") {
			domainOfSite = strings.TrimSuffix(siteFilename, ".toml")
			sites = append(sites, getSiteByDomain(domainOfSite))
		}
	}
	return sites
}

func getChunk(domain string) string {
	chunkString := readFile(path.Join(config.SiteStorageDirectory, "chunks", domain))
	return chunkString
}

func getResolver() string {
	return os.Getenv("NGMAN_PROXY_RESOLVER")
}

func chunkExists(domain string) bool {
	_, err := os.Stat(path.Join(config.SiteStorageDirectory, "chunks", domain))
	return !errors.Is(err, fs.ErrNotExist)
}

func getSiteByDomain(domain string) SiteInfo {
	tomlString := readFile(path.Join(config.SiteStorageDirectory, domain+".toml"))
	var siteinfo SiteInfo
	try(toml.Unmarshal([]byte(tomlString), &siteinfo))
	return siteinfo
}

func siteExists(domain string) bool {
	_, err := os.Stat(path.Join(config.SiteStorageDirectory, domain+".toml"))
	return !os.IsNotExist(err)
}

func initSite(domain string, rootPath string) {
	useWildcard := isSubDomain(domain)
	ensureDirExists(rootPath)
	site := SiteInfo{
		Domain:   domain,
		RootPath: rootPath,
	}
	writeHTTPOnlyNginxConfig(site)
	tryPostRunCommand()
	if !certExists(domain, useWildcard) {
		fmt.Println("No certificate found for " + domain + ". Generating one...")
		tryGenerateCertificate(domain, rootPath, useWildcard)
	}
	if certExists(domain, useWildcard) {
		updateSite(site)
		tryPostRunCommand()
	}
}

func certExists(domain string, useWildcard bool) bool {
	var certFileName string
	if useWildcard {
		certFileName = path.Join(config.CertificateRootPath, getWildcardName(domain)+".crt")
	} else {
		certFileName = path.Join(config.CertificateRootPath, domain+".crt")
	}
	return fileExists(certFileName)
}

func tryGenerateCertificate(domain string, rootPath string, wildcard bool) {
	if config.GenerateCertCommand == "" {
		fmt.Println("No certificate generation command specified. Skipping...")
		return
	}
	if wildcard {
		wildcardName := getWildcardName(domain)
		wildcardName = strings.Replace(wildcardName, "_", "*", 1)
		domain = wildcardName
	}
	commandLine := config.GenerateCertCommand + " " + domain + " " + rootPath
	arguments := []string{"--login", "-c", commandLine}
	fmt.Println("Running certificate generation command: bash " + strings.Join(arguments, " "))
	cmd := exec.Command("bash", arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	try(cmd.Run())
}

func tryPostRunCommand() {
	if config.PostRunCommand != "" {
		cmd := exec.Command("bash", "--login", "-c", config.PostRunCommand)
		// attach stdout and stderr to the current process
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		try(cmd.Run())
	}
}

func updateSite(siteinfo SiteInfo) {
	writeSiteInfo(siteinfo)
	writeNginxConfig(siteinfo)
}

func writeSiteInfo(siteinfo SiteInfo) {
	// marshal the siteinfo to toml
	tomlString := must(toml.Marshal(siteinfo))
	try(writeFile(path.Join(config.SiteStorageDirectory, siteinfo.Domain+".toml"), tomlString))
}
func writeNginxConfig(site SiteInfo) {
	context := RenderContext{
		Site:       site,
		Config:     config,
		SSLEnabled: true,
	}
	renderNginxConfig(site, context)
}

func writeHTTPOnlyNginxConfig(site SiteInfo) {
	context := RenderContext{
		Site:       site,
		Config:     config,
		SSLEnabled: false,
	}
	renderNginxConfig(site, context)
}

func renderNginxConfig(site SiteInfo, context RenderContext) {
	output := renderTemplate(context)
	renderedString := removeRedundantEmptyLines(string(output))
	try(writeFile(path.Join(config.NginxSiteConfigDirectory, site.Domain+".conf"), []byte(renderedString)))
}

func loadConfig() {
	configDir := getConfigDir()
	ensureDirExists(configDir)
	configFilename := path.Join(configDir, "config.toml")
	if !fileExists(configFilename) {
		config = GlobalConfig{
			SiteStorageDirectory:     path.Join(configDir, "sites"),
			TemplateFile:             path.Join(configDir, "nginx.txt"),
			NginxSiteConfigDirectory: path.Join(configDir, "sites-enabled"),
			PostRunCommand:           "",
			GenerateCertCommand:      "",
			CertificateRootPath:      "/ssl/certificates",
			WebRootPath:              "/var/www",
		}

		tomlString := must(toml.Marshal(config))
		try(writeFile(configFilename, tomlString))
	} else {
		tomlString := readFile(configFilename)
		try(toml.Unmarshal([]byte(tomlString), &config))
	}
}

func getConfigDir() string {
	homeDir := must(os.UserHomeDir())
	configDir := path.Join(homeDir, ".ngman")
	return configDir
}

// sub.domain.com -> _.domain.com
func getWildcardName(domain string) string {
	domainParts := strings.Split(domain, ".")
	return "_." + strings.Join(domainParts[len(domainParts)-2:], ".")
}

func renderTemplate(data RenderContext) []byte {
	var buffer bytes.Buffer
	err := rootTemplate.ExecuteTemplate(&buffer, "nginx", data)
	if err != nil {
		log.Fatal(err)
	}
	return buffer.Bytes()
}

func loadTemplates() {
	rootTemplate = template.Must(template.New("nginx").Funcs(template.FuncMap{"getWildcardName": getWildcardName, "getChunk": getChunk, "chunkExists": chunkExists, "getResolver": getResolver}).ParseFiles(config.TemplateFile))
}
