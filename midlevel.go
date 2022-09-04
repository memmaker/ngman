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
	chunkString := readFile(config.SiteStorageDirectory + "/chunks/" + domain)
	return chunkString
}

func chunkExists(domain string) bool {
	_, err := os.Stat(config.SiteStorageDirectory + "/chunks/" + domain)
	return !errors.Is(err, fs.ErrNotExist)
}

func getSiteByDomain(domain string) SiteInfo {
	tomlString := readFile(config.SiteStorageDirectory + "/" + domain + ".toml")
	var siteinfo SiteInfo
	try(toml.Unmarshal([]byte(tomlString), &siteinfo))
	return siteinfo
}

func siteExists(domain string) bool {
	_, err := os.Stat(config.SiteStorageDirectory + "/" + domain + ".toml")
	return !os.IsNotExist(err)
}

func initSite(domain string, useWildcard bool) SiteInfo {
	if !certExists(domain, useWildcard) {
		fmt.Println("No certificate found for " + domain + ". Generating one...")
		tryGenerateCertificate(domain, useWildcard)
	}
	return SiteInfo{
		Domain:          domain,
		UseWildcardCert: useWildcard,
	}
}

func certExists(domain string, useWildcard bool) bool {
	var certFileName string
	if useWildcard {
		certFileName = config.CertificateRootPath + "/" + getWildcardName(domain) + ".crt"
	} else {
		certFileName = config.CertificateRootPath + "/" + domain + ".crt"
	}
	return fileExists(certFileName)
}

func tryGenerateCertificate(domain string, wildcard bool) {
	if config.GenerateCertCommand == "" {
		fmt.Println("No certificate generation command specified. Skipping...")
		return
	}
	if wildcard {
		wildcardName := getWildcardName(domain)
		wildcardName = strings.Replace(wildcardName, "_", "*", 1)
		domain = wildcardName
	}
	command := config.GenerateCertCommand + " " + domain
	cmd := exec.Command("bash", "--login", "-c", command)
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
	try(writeFile(config.SiteStorageDirectory+"/"+siteinfo.Domain+".toml", tomlString))
}
func writeNginxConfig(site SiteInfo) {
	context := RenderContext{
		Site:   site,
		Config: config,
	}
	output := renderTemplate(context)
	renderedString := removeRedundantEmptyLines(string(output))
	try(writeFile(config.NginxSiteConfigDirectory+"/"+site.Domain+".conf", []byte(renderedString)))
}

func loadConfig() {
	homeDir := must(os.UserHomeDir())
	configDir := homeDir + "/.ngman"
	ensureDirExists(configDir)
	configFilename := configDir + "/config.toml"
	if !fileExists(configFilename) {
		config = GlobalConfig{
			SiteStorageDirectory:     configDir + "/sites",
			TemplateFile:             configDir + "/nginx.txt",
			NginxSiteConfigDirectory: configDir + "/sites-enabled",
			PostRunCommand:           "",
			GenerateCertCommand:      "",
			CertificateRootPath:      "/ssl/certificates",
		}

		tomlString := must(toml.Marshal(config))
		try(writeFile(configFilename, tomlString))
	} else {
		tomlString := readFile(configFilename)
		try(toml.Unmarshal([]byte(tomlString), &config))
	}
}

// sub.domain.com -> _.domain.com
func getWildcardName(domain string) string {
	//split string by "."
	domainParts := strings.Split(domain, ".")
	// take the last two parts of the domain and join them with "."
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
	rootTemplate = template.Must(template.New("nginx").Funcs(template.FuncMap{"getWildcardName": getWildcardName, "getChunk": getChunk, "chunkExists": chunkExists}).ParseFiles(config.TemplateFile))
}
