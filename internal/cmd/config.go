package cmd

import "fmt"

type config struct {
	Providers []provider `hcl:"provider,block"`
}

type provider struct {
	Namespace string `hcl:"namespace,label"`
	Name      string `hcl:"name,label"`

	// Sources
	Manual   *manualSource   `hcl:"manual,block"`
	GitHub   *gitHubSource   `hcl:"github,block"`
	Registry *registrySource `hcl:"registry,block"`
}

func (p provider) String() string {
	return fmt.Sprintf("%s/%s", p.Namespace, p.Name)
}

type gitHubSource struct {
	Repository    string `hcl:"repository"`
	PublicKeyFile string `hcl:"public_key_file"`
}

type registrySource struct {
	Source string `hcl:"source"`
}

type manualSource struct {
	// TODO: support a manual source
}

func (conf *config) Validate() error {
	for _, p := range conf.Providers {
		if p.Namespace == "" {
			return fmt.Errorf("a blank namespace is not allowed")
		}
		if p.Name == "" {
			return fmt.Errorf("a blank name is not allowed")
		}

		if p.GitHub == nil && p.Registry == nil && p.Manual == nil {
			return fmt.Errorf("a source block of github, registry, or manual is required for provider %q", p)
		}
	}
	return nil
}
