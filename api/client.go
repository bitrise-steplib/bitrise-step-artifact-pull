package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bitrise-io/go-utils/retry"
)

const (
	bitriseAPIClientTimeout = 30
)

type DefaultBitriseAPIClient struct {
	httpClient *http.Client
	authToken  string
	baseURL    string
}

func NewDefaultBitriseAPIClient(baseURL, authToken string) (DefaultBitriseAPIClient, error) {
	httpClient := retry.NewHTTPClient().StandardClient()
	httpClient.Timeout = time.Second * bitriseAPIClientTimeout

	c := DefaultBitriseAPIClient{
		httpClient: httpClient,
		authToken:  authToken,
		baseURL:    baseURL,
	}

	return c, nil
}

// ListBuildArtifacts gets the list of artifact details for a given build slug (also performs paging and calls the endpoint multiple times if needed)
func (c *DefaultBitriseAPIClient) ListBuildArtifacts(appSlug, buildSlug string) ([]ArtifactListElementResponseModel, error) {
	var artifacts []ArtifactListElementResponseModel
	requestPath := fmt.Sprintf("v0.2/apps/%s/builds/%s/artifacts", appSlug, buildSlug)

	var next string
	for {
		resp, err := c.get(requestPath, next)
		if err != nil {
			return nil, err
		}

		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		responseBodyCloser(resp)
		if err != nil {
			return nil, err
		}

		var responseModel ListBuildArtifactsResponse
		if err := json.Unmarshal(respBody, &responseModel); err != nil {
			return nil, err
		}

		artifacts = append(artifacts, responseModel.Data...)

		if len(responseModel.Paging.Next) > 0 {
			next = responseModel.Paging.Next
		} else {
			break
		}
	}

	return artifacts, nil
}

// ShowBuildArtifact gets the details of a given artifact identified by its slug
func (c *DefaultBitriseAPIClient) ShowBuildArtifact(appSlug, buildSlug, artifactSlug string) (ArtifactResponseItemModel, error) {
	requestPath := fmt.Sprintf("v0.2/apps/%s/builds/%s/artifacts/%s", appSlug, buildSlug, artifactSlug)

	resp, err := c.get(requestPath, "") //nolint: bodyclose
	if err != nil {
		return ArtifactResponseItemModel{}, err
	}
	defer responseBodyCloser(resp)

	var respBody []byte
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArtifactResponseItemModel{}, err
	}

	var responseModel ShowBuildArtifactResponse
	if err := json.Unmarshal(respBody, &responseModel); err != nil {
		return ArtifactResponseItemModel{}, err
	}

	return responseModel.Data, nil
}

func (c DefaultBitriseAPIClient) get(endpoint, next string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.authToken)

	if next != "" {
		queryValues := req.URL.Query()
		queryValues.Add("next", next)
		req.URL.RawQuery = queryValues.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		err = fmt.Errorf("request to %s failed - status code should be 2XX (%d)", req.URL.String(), resp.StatusCode)
	}

	return resp, err
}

func responseBodyCloser(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf(" [!] Failed to close response body: %+v", err)
	}
}
