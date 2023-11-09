package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	tfjson "github.com/hashicorp/terraform-json"
)

// https://developer.hashicorp.com/terraform/internals/json-format#plan-representation
func parsePlan(ctx context.Context, client *http.Client, url, token string) (*tfjson.Plan, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Unexpected status was returned: %d", resp.StatusCode)
	}

	var plan *tfjson.Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, err
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return plan, nil
}
