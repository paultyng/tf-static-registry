package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type shasum struct {
	Sum  string
	File string
}

func downloadSHASUMS(ctx context.Context, client *http.Client, url string) ([]shasum, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to GET SHASUMS file: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read SHASUMS body: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(body))
	sums := []shasum{}
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid SHASUMS line: %q", line)
		}

		sums = append(sums, shasum{
			Sum:  fields[0],
			File: fields[1],
		})
	}

	return sums, nil
}
