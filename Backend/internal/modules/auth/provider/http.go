package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func postForm(ctx context.Context, client *http.Client, endpoint string, form url.Values, result any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	return doJSON(client, request, result)
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, accessToken string, result any) error {
	_, err := getJSONWithHeaders(ctx, client, endpoint, accessToken, result)
	return err
}

func getJSONWithHeaders(ctx context.Context, client *http.Client, endpoint string, accessToken string, result any) (http.Header, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(accessToken) != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	request.Header.Set("Accept", "application/json")
	return doJSONWithHeaders(client, request, result)
}

func doJSON(client *http.Client, request *http.Request, result any) error {
	_, err := doJSONWithHeaders(client, request, result)
	return err
}

func doJSONWithHeaders(client *http.Client, request *http.Request, result any) (http.Header, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("OAuth endpoint returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return nil, err
	}
	return response.Header.Clone(), nil
}
