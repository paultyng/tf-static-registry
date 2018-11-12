package reggen

type wellKnownServices struct {
	ModulesV1   string `json:"modules.v1"`
	ProvidersV1 string `json:"providers.v1"`
}

type moduleProviderDep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type moduleDep struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

type moduleVersions struct {
	Modules []*moduleProviderVersions `json:"modules"`
}

type moduleProviderVersions struct {
	Source   string           `json:"source"`
	Versions []*moduleVersion `json:"versions"`
}
type moduleVersion struct {
	Version    string              `json:"version"`
	Root       versionSubmodule    `json:"root"`
	Submodules []*versionSubmodule `json:"submodules"`
}

type versionSubmodule struct {
	Path         string               `json:"path,omitempty"`
	Providers    []*moduleProviderDep `json:"providers"`
	Dependencies []*moduleDep         `json:"dependencies"`
}
