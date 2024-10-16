package jsonschema

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

// LoadMap fetches the JSON schema from a given URL
func LoadMap(ctx context.Context, url string) (map[string]any, error) {
	var err error
	var body []byte

	if strings.HasPrefix(url, "http") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		body, err = os.ReadFile(url)
		if err != nil {
			return nil, err
		}
	}

	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, err
	}

	return schema, nil
}
