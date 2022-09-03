package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"text/template"
)

var config GlobalConfig

type GlobalConfig struct {
	CertificateRootPath      string
	SiteStorageDirectory     string
	NginxSiteConfigDirectory string
	TemplateFile             string
	PostRunCommand           string
	GenerateCertCommand      string
}

type SiteInfo struct {
	Domain          string
	UseWildcardCert bool
	UsePHP          bool
	UseChunk        bool
	Rewrites        []string
	StaticLocations []StaticLocation
	ProxyLocations  []ReverseProxyLocation
	MiscOptions     []string
	RootPath        string
}

type ReverseProxyLocation struct {
	URLLocation string
	Endpoint    string
	Headers     map[string]string
}

type StaticLocation struct {
	URLLocation string
	RootPath    string
}

type RenderContext struct {
	Site   SiteInfo
	Config GlobalConfig
}

var rootTemplate *template.Template

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
		if !strings.HasSuffix(siteFilename, ".json") {
			continue
		}
		domainOfSite := strings.TrimSuffix(siteFilename, ".json")
		sites = append(sites, getSiteByDomain(domainOfSite))
	}
	return sites
}

func writeAll() {
	sites := getAllSites()
	for _, site := range sites {
		writeNginxConfig(site)
	}
}

func getChunk(domain string) string {
	chunkString := readFile(config.SiteStorageDirectory + "/chunks/" + domain)
	return chunkString
}

func getSiteByDomain(domain string) SiteInfo {
	jsonString := readFile(config.SiteStorageDirectory + "/" + domain + ".json")
	var siteinfo SiteInfo
	try(json.Unmarshal([]byte(jsonString), &siteinfo))
	return siteinfo
}

func siteExists(domain string) bool {
	_, err := os.Stat(config.SiteStorageDirectory + "/" + domain + ".json")
	return !os.IsNotExist(err)
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

func initSite(domain string, useWildcard bool) SiteInfo {
	if !certExists(domain, useWildcard) {
		fmt.Println("No certificate found for " + domain + ". Generating one...")
		generateCertificate(domain, useWildcard)
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

func generateCertificate(domain string, wildcard bool) {
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

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

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

func tryPostRunCommand() {
	if config.PostRunCommand != "" {
		cmd := exec.Command("bash", "--login", "-c", config.PostRunCommand)
		// attach stdout and stderr to the current process
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		try(cmd.Run())
	}
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

func deleteSite(domain string) {
	try(os.Remove(config.SiteStorageDirectory + "/" + domain + ".json"))
	try(os.Remove(config.NginxSiteConfigDirectory + "/" + domain + ".conf"))
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

func updateSite(siteinfo SiteInfo) {
	writeSiteInfo(siteinfo)
	writeNginxConfig(siteinfo)
}

func writeSiteInfo(siteinfo SiteInfo) {
	// marshal the siteinfo to json
	jsonString := must(json.MarshalIndent(siteinfo, "", "  "))
	try(writeFile(config.SiteStorageDirectory+"/"+siteinfo.Domain+".json", jsonString))
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

func removeRedundantEmptyLines(content string) string {
	return strings.Replace(content, "\n\n", "\n", -1)
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
	rootTemplate = template.Must(template.New("nginx").Funcs(template.FuncMap{"getWildcardName": getWildcardName, "getChunk": getChunk}).ParseFiles(config.TemplateFile))
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("[isRoot] Unable to get current user: %s", err)
	}
	return currentUser.Username == "root"
}

func loadConfig() {
	homeDir := must(os.UserHomeDir())
	configDir := homeDir + "/.ngman"
	ensureDirExists(configDir)
	configFilename := configDir + "/config.json"
	if !fileExists(configFilename) {
		config = GlobalConfig{
			CertificateRootPath:      "/ssl/certificates",
			SiteStorageDirectory:     configDir + "/sites",
			NginxSiteConfigDirectory: "/etc/nginx/sites-enabled",
			TemplateFile:             configDir + "/nginx.txt",
			PostRunCommand:           "service nginx reload",
			GenerateCertCommand:      "create_ssl_cert",
		}
		// to json
		jsonString := must(json.MarshalIndent(config, "", "  "))
		try(writeFile(configFilename, jsonString))
	} else {
		jsonString := readFile(configFilename)
		try(json.Unmarshal([]byte(jsonString), &config))
	}
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

func ValueOrDefault[T comparable](value T, defaultValue T) T {
	if value == *new(T) {
		return defaultValue
	}
	return value
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

// sub.domain.com -> _.domain.com
func getWildcardName(domain string) string {
	//split string by "."
	domainParts := strings.Split(domain, ".")
	// take the last two parts of the domain and join them with "."
	return "_." + strings.Join(domainParts[len(domainParts)-2:], ".")
}
