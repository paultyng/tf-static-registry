package cmd

var (
	// 5.0 is the only protocol used in registries currently
	providerProtocols = []string{"5.0"}
)

type wellKnownTerraform struct {
	ProvidersV1 string `json:"providers.v1"`
}

type registryData struct {
	ProviderVersions map[providerVersionsKey]providerVersionsIndex
	Downloads        map[providerDownloadKey]providerDownloadIndex
}

type providerVersionsKey struct {
	Namespace string
	Name      string
}

type providerVersionsIndex struct {
	ID       string            `json:"id"`
	Warnings []string          `json:"warnings"`
	Versions []providerVersion `json:"versions"`
}

type providerVersion struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []platform `json:"platforms"`
}

type platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type providerDownloadKey struct {
	Namespace string
	Name      string
	Version   string
	OS        string
	Arch      string
}

type providerDownloadIndex struct {
	Protocols           []string    `json:"protocols"`
	OS                  string      `json:"os"`
	Arch                string      `json:"arch"`
	Filename            string      `json:"filename"`
	DownloadURL         string      `json:"download_url"`
	ShasumsURL          string      `json:"shasums_url"`
	ShasumsSignatureURL string      `json:"shasums_signature_url"`
	Shasum              string      `json:"shasum"`
	SigningKeys         signingKeys `json:"signing_keys"`
}

type signingKeys struct {
	GPGPublicKeys []gpgPublicKey `json:"gpg_public_keys"`
}

type gpgPublicKey struct {
	KeyID          string `json:"key_id"`
	ASCIIArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceURL      string `json:"source_url"`
}
