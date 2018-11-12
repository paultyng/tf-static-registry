package reggen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const terraformGetHeader = "X-Terraform-Get"

var fileNameRegexp = regexp.MustCompile(
	`^` + // Beginning
		`terraform` + // Always starts with `terraform`
		`(-(?P<provider>[a-zA-Z0-9]|[a-zA-Z0-9][_.a-zA-Z0-9]*[a-zA-Z0-9]))` + // Necessary provider
		`-(?P<name>[a-zA-Z0-9]|[a-zA-Z0-9][-_.a-zA-Z0-9]*[a-zA-Z0-9])` + // Required name
		`-(?P<version>\d+\.\d+\.\d+)` + // Required version
		`\.(tar\.gz|tgz)` + // tar.gz extension
		`$`)

func Generate(modulepath, outpath string) error {
	g := &generator{
		outpath:    outpath,
		modulepath: modulepath,
		services: wellKnownServices{
			ModulesV1:   "/v1/modules/",
			ProvidersV1: "/v1/providers/",
		},
	}

	// TODO: strategies for S3 vs Caddy, write out Caddyfile, etc

	err := os.RemoveAll(g.outpath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(g.outpath, 0775)
	if err != nil {
		return err
	}

	err = g.CollectModules()
	if err != nil {
		return err
	}

	err = g.WriteJSONFile(".well-known/terraform.json", g.services)
	if err != nil {
		return err
	}

	err = g.WriteLatestVersions()
	if err != nil {
		return err
	}

	err = g.WriteDownloads()
	if err != nil {
		return err
	}

	return nil
}

type collectedModuleVersion struct {
	md5       string
	namespace string
	name      string
	provider  string
	version   string
	src       string
	download  string
}

type generator struct {
	outpath    string
	modulepath string

	services wellKnownServices

	moduleVersions []collectedModuleVersion
}

func (g *generator) CollectModules() error {
	namespaces, err := ioutil.ReadDir(g.modulepath)
	if err != nil {
		return err
	}

	for _, nsdir := range namespaces {
		if !nsdir.IsDir() {
			continue
		}
		ns := nsdir.Name()

		nspath := filepath.Join(g.modulepath, nsdir.Name())
		files, err := ioutil.ReadDir(nspath)
		if err != nil {
			return err
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			match := fileNameRegexp.FindStringSubmatch(f.Name())
			if match == nil {
				continue
			}

			file := filepath.Join(nspath, f.Name())
			md5, err := fileMD5(file)
			if err != nil {
				return err
			}

			captures := make(map[string]string)
			for i, name := range fileNameRegexp.SubexpNames() {
				if i != 0 {
					captures[name] = match[i]
				}
			}

			provider := captures["provider"]
			name := captures["name"]
			version := captures["version"]

			downloadFile := filepath.Join(g.outpath, "downloads", fmt.Sprintf("%s.tar.gz", md5))
			log.Printf("copying download file from %s/%s", ns, f.Name())
			err = copy(file, downloadFile)
			if err != nil {
				return err
			}

			downloadURL, err := filepath.Rel(g.outpath, downloadFile)
			if err != nil {
				return err
			}

			g.moduleVersions = append(g.moduleVersions, collectedModuleVersion{
				md5:       md5,
				namespace: ns,
				name:      name,
				provider:  provider,
				version:   version,
				src:       file,
				download:  path.Join("/", downloadURL),
			})

			log.Printf("ingesting %s", f.Name())
		}
	}

	return nil
}

func (g *generator) WriteJSONFile(name string, v interface{}) error {
	log.Printf("creating file %s", name)
	filename := filepath.Join(g.outpath, name)
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (g *generator) WriteLatestVersions() error {
	type module struct {
		namespace string
		name      string
		provider  string
	}

	modules := map[module][]string{}

	for _, mv := range g.moduleVersions {
		m := module{
			namespace: mv.namespace,
			name:      mv.name,
			provider:  mv.provider,
		}
		versions := modules[m]
		if len(versions) == 0 {
			versions = []string{}
			modules[m] = versions
		}

		modules[m] = append(versions, mv.version)
	}

	for mod, versions := range modules {
		path := filepath.Join("v1", "modules", mod.namespace, mod.name, mod.provider, "versions", "index.json")

		mvs := make([]*moduleVersion, 0, len(versions))

		for _, v := range versions {
			mvs = append(mvs, &moduleVersion{
				Version: v,
			})
		}

		latestVersions := &moduleVersions{
			Modules: []*moduleProviderVersions{
				&moduleProviderVersions{
					Source:   fmt.Sprintf("%s/%s/%s", mod.namespace, mod.name, mod.provider),
					Versions: mvs,
				},
			},
		}

		err := g.WriteJSONFile(path, latestVersions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) WriteDownloads() error {
	// this is Caddy specific
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

	for _, mv := range g.moduleVersions {
		url := path.Join("/v1", "modules", mv.namespace, mv.name, mv.provider, mv.version, "download")

		// this is go-getter syntax: https://github.com/hashicorp/go-getter#url-format
		goGetterURL := "https://registry.lvh.me:2015" + path.Join("/", mv.download) + "//*?archive=tar.gz"

		fmt.Fprintf(f, "header %s %s %s\n", url, terraformGetHeader, goGetterURL)
		fmt.Fprintf(f, "status 204 %s\n", url)
	}

	return nil
}
