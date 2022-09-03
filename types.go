package main

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
