package reggen

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
)

const terraformGetHeader = "X-Terraform-Get"

var moduleFileRegexp = regexp.MustCompile(
	`^` + // Beginning
		`(terraform-)?` + // Can start with `terraform-`
		`(?P<provider>[a-zA-Z0-9]|[a-zA-Z0-9][_.a-zA-Z0-9]*[a-zA-Z0-9])` + // Necessary provider
		`-(?P<name>[a-zA-Z0-9]|[a-zA-Z0-9][-_.a-zA-Z0-9]*[a-zA-Z0-9])` + // Required name
		`-(?P<version>\d+\.\d+\.\d+)` + // Required version
		`\.(tar\.gz|tgz)` + // tar.gz extension
		`$`)

type generator interface {
	Setup([]CollectedModuleVersion) error
	WriteLatestVersions([]CollectedModuleVersion) error
	WriteDownloads([]CollectedModuleVersion) error
}

func Generate(modulepath, outpath string) error {
	var g generator

	log.Printf("creating static registry for %s", "Caddy")

	g = &caddyGenerator{
		fileSystemGenerator: fileSystemGenerator{
			outpath:    outpath,
			modulepath: modulepath,
			services: wellKnownServices{
				ModulesV1:   "/v1/modules/",
				ProvidersV1: "/v1/providers/",
			},
		},
	}

	// TODO: strategies for S3 vs Caddy, write out Caddyfile, etc

	collected, err := collectModules(modulepath)
	if err != nil {
		return err
	}

	err = g.Setup(collected)
	if err != nil {
		return err
	}

	err = g.WriteLatestVersions(collected)
	if err != nil {
		return err
	}

	err = g.WriteDownloads(collected)
	if err != nil {
		return err
	}

	return nil
}

type CollectedModuleVersion struct {
	MD5       string
	Namespace string
	Name      string
	Provider  string
	Version   string
	Src       string
}

func collectModules(modulepath string) ([]CollectedModuleVersion, error) {
	namespaces, err := ioutil.ReadDir(modulepath)
	if err != nil {
		return nil, err
	}

	var moduleVersions []CollectedModuleVersion

	for _, nsdir := range namespaces {
		if !nsdir.IsDir() {
			continue
		}
		ns := nsdir.Name()

		nspath := filepath.Join(modulepath, nsdir.Name())
		files, err := ioutil.ReadDir(nspath)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			match := moduleFileRegexp.FindStringSubmatch(f.Name())
			if match == nil {
				continue
			}

			file := filepath.Join(nspath, f.Name())
			md5, err := fileMD5(file)
			if err != nil {
				return nil, err
			}

			captures := make(map[string]string)
			for i, name := range moduleFileRegexp.SubexpNames() {
				if i != 0 {
					captures[name] = match[i]
				}
			}

			provider := captures["provider"]
			name := captures["name"]
			version := captures["version"]

			moduleVersions = append(moduleVersions, CollectedModuleVersion{
				MD5:       md5,
				Namespace: ns,
				Name:      name,
				Provider:  provider,
				Version:   version,
				Src:       file,
			})

			log.Printf("ingesting %s", f.Name())
		}
	}

	return moduleVersions, nil
}
