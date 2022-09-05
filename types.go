package main

type GlobalConfig struct {
	CertificateRootPath      string
	SiteStorageDirectory     string
	WebRootPath              string
	NginxSiteConfigDirectory string
	TemplateFile             string
	PostRunCommand           string
	GenerateCertCommand      string
}

type SiteInfo struct {
	Domain          string
	UsePHP          bool
	StaticLocations []StaticLocation
	ProxyLocations  []ReverseProxyLocation
	MiscOptions     []string
	RootPath        string
}

func (s SiteInfo) UseWildcardCert() bool {
	return isSubDomain(s.Domain)
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
