package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/shurcooL/githubv4"
	"golang.org/x/crypto/openpgp"
)

func (cmd *generateCmd) collectGitHubProvider(ctx context.Context, p provider, rd registryData) error {
	cmd.ui.Info(fmt.Sprintf("\t[%q] collecting GitHub information...", p))

	if cmd.githubClient == nil {
		return fmt.Errorf("no GitHub client configured, please specify api token")
	}

	repoParts := strings.Split(p.GitHub.Repository, "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("malformed github repository %q", p.GitHub.Repository)
	}
	owner, name := repoParts[0], repoParts[1]

	if p.GitHub.PublicKeyFile == "" {
		return fmt.Errorf("a public key file is required")
	}

	keyRingData, err := ioutil.ReadFile(p.GitHub.PublicKeyFile)
	if err != nil {
		return fmt.Errorf("unable to read public key file %q: %w", p.GitHub.PublicKeyFile, err)
	}

	keyRing, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(keyRingData))
	if err != nil {
		return fmt.Errorf("unable to read armored key ring for %q: %w", p.GitHub.PublicKeyFile, err)
	}
	if len(keyRing) != 1 {
		return fmt.Errorf("expected 1 key in %q, got %d", p.GitHub.PublicKeyFile, len(keyRing))
	}

	key := keyRing[0]

	type pageInfo struct {
		EndCursor   githubv4.String
		HasNextPage bool
	}

	type releaseAsset struct {
		ContentType string
		DownloadURL string `graphql:"downloadUrl"`
		Name        string
	}

	type release struct {
		TagName       string
		IsPrerelease  bool
		IsDraft       bool
		ReleaseAssets struct {
			PageInfo pageInfo
			Nodes    []releaseAsset
		} `graphql:"releaseAssets(first: 100)"`
	}

	var q struct {
		Repository struct {
			Releases struct {
				PageInfo pageInfo
				Nodes    []release
			} `graphql:"releases(first: 100, orderBy: { field: CREATED_AT, direction: DESC }, after: $releasesCursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner":          githubv4.String(owner),
		"name":           githubv4.String(name),
		"releasesCursor": (*githubv4.String)(nil),
	}

	versionsIndex := providerVersionsIndex{
		ID:       fmt.Sprintf("%s/%s", p.Namespace, p.Name),
		Warnings: []string{},
	}

	for {
		err := cmd.githubClient.Query(ctx, &q, variables)
		if err != nil {
			return err
		}

		// check if any releases have multiple pages of assets, not yet supported...
		for _, r := range q.Repository.Releases.Nodes {
			// declaration up here since we are using goto
			var (
				sumsAsset    *releaseAsset
				sigAsset     *releaseAsset
				platforms    []platform
				assetsByName = map[string]releaseAsset{}
				sums         []shasum
			)

			cmd.ui.Info(fmt.Sprintf("\t[%q] processing tag %q...", p, r.TagName))

			if r.ReleaseAssets.PageInfo.HasNextPage {
				return fmt.Errorf("release %q has over 100 assets, this is not yet supported", r.TagName)
			}
			ver := strings.TrimPrefix(r.TagName, "v")
			if _, err := version.NewSemver(ver); err != nil {
				cmd.ui.Warn(fmt.Sprintf("\t\t[%q] skipping %q, not valid semver: %s", p, r.TagName, err))
				goto NextRelease
			}
			if l := len(r.ReleaseAssets.Nodes); l == 0 {
				cmd.ui.Warn(fmt.Sprintf("\t\t[%q] skipping %q, no release assets", p, r.TagName))
				goto NextRelease
			}

			for _, ra := range r.ReleaseAssets.Nodes {
				ra := ra
				switch {
				case strings.HasSuffix(ra.Name, "_SHA256SUMS"):
					sumsAsset = &ra
					continue
				case strings.HasSuffix(ra.Name, "_SHA256SUMS.sig"):
					sigAsset = &ra
					continue
				}
				assetsByName[ra.Name] = ra
			}
			if sumsAsset == nil {
				cmd.ui.Warn(fmt.Sprintf("\t[%q] skipping %q, no SHASUMS asset found", p, r.TagName))
				goto NextRelease
			}
			if sigAsset == nil {
				cmd.ui.Warn(fmt.Sprintf("\t[%q] skipping %q, no signature asset found", p, r.TagName))
				goto NextRelease
			}

			sums, err = downloadSHASUMS(ctx, cmd.httpClient, sumsAsset.DownloadURL)
			if err != nil {
				cmd.ui.Warn(fmt.Sprintf("\t[%q] skipping %q, unable to download SHASUMS asset: %s", p, r.TagName, err))
				goto NextRelease
			}

			for _, sum := range sums {
				ra, ok := assetsByName[sum.File]
				if !ok {
					cmd.ui.Warn(fmt.Sprintf("\t[%q] skipping %q, file referenced by SHASUMS not found in release assets: %q", p, r.TagName, sum.File))
					goto NextRelease
				}
				name := sum.File
				name = strings.TrimSuffix(name, path.Ext(name))
				nameParts := strings.Split(name, "_")
				if len(nameParts) != 4 {
					cmd.ui.Warn(fmt.Sprintf("\t[%q] skipping %q, malformed asset file: %q", p, r.TagName, ra.Name))
					goto NextRelease
				}
				os, arch := nameParts[2], nameParts[3]

				platforms = append(platforms, platform{
					OS:   os,
					Arch: arch,
				})

				rd.Downloads[providerDownloadKey{
					Namespace: p.Namespace,
					Name:      p.Name,

					Version: ver,

					OS:   os,
					Arch: arch,
				}] = providerDownloadIndex{
					OS:   os,
					Arch: arch,

					Filename: sum.File,
					// TODO: support local copy of bytes?
					DownloadURL:         ra.DownloadURL,
					Shasum:              sum.Sum,
					ShasumsURL:          sumsAsset.DownloadURL,
					ShasumsSignatureURL: sigAsset.DownloadURL,

					SigningKeys: signingKeys{
						GPGPublicKeys: []gpgPublicKey{
							{
								KeyID:      key.PrimaryKey.KeyIdString(),
								ASCIIArmor: string(keyRingData),

								// currently only the HashiCorp registry supports trust signatures
								TrustSignature: "",
							},
						},
					},

					Protocols: providerProtocols,
				}
			}

			versionsIndex.Versions = append(versionsIndex.Versions, providerVersion{
				Version:   ver,
				Platforms: platforms,

				Protocols: providerProtocols,
			})
		NextRelease:
		}

		if !q.Repository.Releases.PageInfo.HasNextPage {
			break
		}
		variables["releasesCursor"] = githubv4.NewString(q.Repository.Releases.PageInfo.EndCursor)
	}

	rd.ProviderVersions[providerVersionsKey{
		Namespace: p.Namespace,
		Name:      p.Name,
	}] = versionsIndex

	// TODO: ui done?

	return nil
}
