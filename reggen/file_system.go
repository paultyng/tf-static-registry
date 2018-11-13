package reggen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type fileSystemGenerator struct {
	outpath    string
	modulepath string

	services wellKnownServices
}

func (g *fileSystemGenerator) Setup(moduleVersions []CollectedModuleVersion) error {
	err := os.RemoveAll(g.outpath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(g.outpath, 0775)
	if err != nil {
		return err
	}

	err = g.WriteJSONFile(".well-known/terraform.json", g.services)
	if err != nil {
		return err
	}

	return nil
}

func (g *fileSystemGenerator) WriteJSONFile(name string, v interface{}) error {
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

func (g *fileSystemGenerator) WriteLatestVersions(collectedModuleVersions []CollectedModuleVersion) error {
	type module struct {
		namespace string
		name      string
		provider  string
	}

	modules := map[module][]string{}

	for _, mv := range collectedModuleVersions {
		m := module{
			namespace: mv.Namespace,
			name:      mv.Name,
			provider:  mv.Provider,
		}
		versions := modules[m]
		if len(versions) == 0 {
			versions = []string{}
			modules[m] = versions
		}

		modules[m] = append(versions, mv.Version)
	}

	for mod, versions := range modules {
		path := filepath.Join(g.services.ModulesV1, mod.namespace, mod.name, mod.provider, "versions", "index.json")

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

func (g *fileSystemGenerator) writeDownload(mv CollectedModuleVersion) (string, error) {
	downloadFile := filepath.Join(g.outpath, "downloads", fmt.Sprintf("%s.tar.gz", mv.MD5))
	log.Printf("copying download file for %s/%s/%s/%s", mv.Namespace, mv.Name, mv.Provider, mv.Version)

	err := copy(mv.Src, downloadFile)
	if err != nil {
		return "", err
	}

	downloadURL, err := filepath.Rel(g.outpath, downloadFile)
	if err != nil {
		return "", err
	}

	return downloadURL, nil
}
