package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (cmd *generateCmd) generateNetlify(ctx context.Context, rd registryData) error {
	cmd.ui.Info("\t[netlify] writing service discovery file...")
	err := writeJSONFile(filepath.Join(cmd.outputDir, ".well-known/terraform.json"), wellKnownTerraform{
		ProvidersV1: "/providers/v1/",
	})
	if err != nil {
		return fmt.Errorf("unable to write service discovery file: %w", err)
	}

	cmd.ui.Info("\t[netlify] writing provider version files")
	for k, v := range rd.ProviderVersions {
		err = writeJSONFile(filepath.Join(
			cmd.outputDir,
			"providers/v1",
			strings.ToLower(k.Namespace), strings.ToLower(k.Name),
			"versions.json",
		), v)
		if err != nil {
			return err
		}
	}

	cmd.ui.Info("\t[netlify] writing provider version download files...")
	for k, v := range rd.Downloads {
		err = writeJSONFile(filepath.Join(
			cmd.outputDir,
			"providers/v1",
			strings.ToLower(k.Namespace), strings.ToLower(k.Name),
			fmt.Sprintf("%s-%s-%s.json", k.Version, k.OS, k.Arch),
		), v)
	}

	cmd.ui.Info("\t[netlify] writing redirects file...")
	err = ioutil.WriteFile(
		filepath.Join(
			cmd.outputDir,
			"_redirects",
		),
		[]byte(`
# redirect the individual version requests
/providers/v1/:namespace/:name/:version/download/:os/:arch	/providers/v1/:namespace/:name/:version-:os-:arch.json	200

# redirect the versions list request
/providers/v1/:namespace/:name/versions	/providers/v1/:namespace/:name/versions.json	200
`),
		0644,
	)

	return nil
}

func writeJSONFile(file string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal JSON to write to file %q: %w", file, err)
	}
	dir := filepath.Dir(file)
	err = os.MkdirAll(dir, 0644)
	if err != nil {
		return fmt.Errorf("unable to make directory %q: %w", dir, err)
	}
	err = ioutil.WriteFile(file, bytes, 0644)
	if err != nil {
		return fmt.Errorf("unable to write file %q: %w", file, err)
	}
	return nil
}
