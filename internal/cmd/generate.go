package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type generateCmd struct {
	commonCmd

	serverType string
	outputDir  string

	httpClient   *http.Client
	githubClient *githubv4.Client
}

func (cmd *generateCmd) Synopsis() string {
	return "generates a static registry"
}

func (cmd *generateCmd) Help() string {
	return `Usage: tfstaticregistry generate`
}

func (cmd *generateCmd) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.StringVar(&cmd.serverType, "server", "", "type of server for the registry")
	fs.StringVar(&cmd.outputDir, "output", "", "output directory for static site")
	return fs
}

func (cmd *generateCmd) Run(args []string) int {
	fs := cmd.Flags()
	err := fs.Parse(args)
	if err != nil {
		cmd.ui.Error(fmt.Sprintf("unable to parse flags: %s", err))
		return 1
	}

	return cmd.run(cmd.runInternal)
}

func (cmd *generateCmd) runInternal() error {
	cmd.ui.Info("")
	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(cmd.outputDir)
	cmd.outputDir, err = filepath.Rel(cwd, abs)
	if err != nil {
		return err
	}

	var conf config
	err = hclsimple.DecodeFile("registry.hcl", nil, &conf)
	if err != nil {
		return err
	}
	err = conf.Validate()
	if err != nil {
		return err
	}

	if cmd.serverType == "" {
		if _, err := os.Stat(filepath.Join(cmd.outputDir, ".netlify")); err == nil {
			cmd.serverType = "netlify"
		}
	}

	cmd.httpClient = cleanhttp.DefaultClient()

	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		httpClient := oauth2.NewClient(ctx, src)

		cmd.githubClient = githubv4.NewClient(httpClient)
	}

	cmd.ui.Info(fmt.Sprintf("Output dir:\t%s\nServer type:\t%s\nGitHub:\t\t%t", cmd.outputDir, cmd.serverType, cmd.githubClient != nil))

	r := registryData{
		ProviderVersions: map[providerVersionsKey]providerVersionsIndex{},
		Downloads:        map[providerDownloadKey]providerDownloadIndex{},
	}

	cmd.ui.Info("\nProcessing providers...\n")

	for _, p := range conf.Providers {
		switch {
		case p.GitHub != nil:
			err = cmd.collectGitHubProvider(ctx, p, r)
			if err != nil {
				return fmt.Errorf("unable to collection information for %q: %w", p, err)
			}
		case p.Registry != nil:
			return fmt.Errorf("registry source is not yet supported for %q", p)
		case p.Manual != nil:
			return fmt.Errorf("manual source is not yet supported for %q", p)
		}

	}

	cmd.ui.Info("\nGenerating registry...\n")

	switch cmd.serverType {
	case "netlify":
		err = cmd.generateNetlify(ctx, r)
		if err != nil {
			return fmt.Errorf("unable to generate netlify server: %w", err)
		}
	default:
		return fmt.Errorf("server type %q not supported", cmd.serverType)
	}

	cmd.ui.Info("\nComplete!\n")

	return nil
}
