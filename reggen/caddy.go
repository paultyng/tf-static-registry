package reggen

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

type caddyGenerator struct {
	fileSystemGenerator
}

func (g *caddyGenerator) Setup(moduleVersions []CollectedModuleVersion) error {
	err := g.fileSystemGenerator.Setup(moduleVersions)
	if err != nil {
		return err
	}

	caddyFile := filepath.Join(g.outpath, "Caddyfile")
	log.Println("creating Caddyfile")

	f, err := os.Create(caddyFile)
	if err != nil {
		return err
	}

	fmt.Fprintf(f, `
# replace this with whatever host name you want to use
registry.lvh.me

# registries must use TLS
# generated with mkcert
tls registry.lvh.me.pem registry.lvh.me-key.pem

# setup logging as you see fit or any other Caddyfile additions
log stdout

# block requests to Caddyfile, secrets, and imports
status 404 /Caddyfile
status 404 /caddy-imports
status 404 /secrets

# import additional generated rules
import ./caddy-imports/downloads

# default documents are assumed to be index.json
rewrite {
	to {path} {path}/index.json
}
# all content is written to the "public" directory
root .
`)

	return nil
}

func (g *caddyGenerator) WriteDownloads(moduleVersions []CollectedModuleVersion) error {
	importFile := filepath.Join(g.outpath, "caddy-imports", "downloads")
	log.Printf("creating file caddy-imports/downloads")

	err := os.MkdirAll(filepath.Dir(importFile), 0775)
	if err != nil {
		return err
	}

	f, err := os.Create(importFile)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, mv := range moduleVersions {
		downloadURL, err := g.fileSystemGenerator.writeDownload(mv)
		if err != nil {
			return err
		}

		url := path.Join(g.services.ModulesV1, mv.Namespace, mv.Name, mv.Provider, mv.Version, "download")

		// this is go-getter syntax: https://github.com/hashicorp/go-getter#url-format
		goGetterURL := "https://registry.lvh.me:2015" + path.Join("/", downloadURL) + "//*?archive=tar.gz"

		fmt.Fprintf(f, "header %s %s %s\n", url, terraformGetHeader, goGetterURL)
		fmt.Fprintf(f, "status 204 %s\n", url)
	}

	return nil
}
