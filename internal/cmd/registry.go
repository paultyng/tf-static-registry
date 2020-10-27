package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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

func (cmd *generateCmd) collectRegistryProvider(ctx context.Context, p provider, rd registryData) error {
	cmd.ui.Info(fmt.Sprintf("\t[%q] collecting Registry information...", p))

	parts := strings.Split(p.Registry.Source, "/")
	if len(parts) == 2 {
		parts = append([]string{"registry.terraform.io"}, parts...)
	}
	if len(parts) != 3 {
		return fmt.Errorf("malformed registry source: %q", p.Registry.Source)
	}

	host, namespace, name := parts[0], parts[1], parts[2]

	cmd.ui.Info(fmt.Sprintf("\t[%q] fetching service discovery information...", p))

	var wk wellKnownTerraform
	err := getJSON(ctx, cmd.httpClient, fmt.Sprintf("https://%s/.well-known/terraform.json", host), &wk)
	if err != nil {
		return fmt.Errorf("unable to get service discovery information: %w", err)
	}

	cmd.ui.Info(fmt.Sprintf("\t[%q] fetching versions index...", p))

	var versions providerVersionsIndex
	err = getJSON(ctx, cmd.httpClient,
		fmt.Sprintf("https://%s/%s/%s/%s/versions",
			host,
			wk.ProvidersV1,
			strings.ToLower(namespace),
			strings.ToLower(name),
		), &versions)
	if err != nil {
		return fmt.Errorf("unable to get versions index: %w", err)
	}

	rd.ProviderVersions[providerVersionsKey{
		Namespace: namespace,
		Name:      name,
	}] = versions

	for _, v := range versions.Versions {
		cmd.ui.Info(fmt.Sprintf("\t[%q] fetching version %q...", p, v.Version))
		for _, plat := range v.Platforms {
			var downloadIndex providerDownloadIndex
			err := getJSON(ctx, cmd.httpClient,
				fmt.Sprintf("https://%s/%s/%s/%s/%s/download/%s/%s",
					host,
					wk.ProvidersV1,
					strings.ToLower(namespace),
					strings.ToLower(name),
					v.Version,
					plat.OS,
					plat.Arch,
				), &downloadIndex)
			if err != nil {
				return fmt.Errorf("unable to get download info for %q \"%s/%s\": %w", v.Version, plat.OS, plat.Arch, err)
			}

			rd.Downloads[providerDownloadKey{
				Namespace: namespace,
				Name:      name,
				Version:   v.Version,
				OS:        plat.OS,
				Arch:      plat.Arch,
			}] = downloadIndex
		}
	}

	return nil
}

func getJSON(ctx context.Context, client *http.Client, url string, data interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("unable to GET JSON file: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read JSON body: %w", err)
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		return fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return nil
}
