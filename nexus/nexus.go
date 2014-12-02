package nexus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	createrepo struct {
		Data Data `json:"data"`
	}

	Data struct {
		ContentResourceURI string `json:"contentResourceURI"`
		Id                 string `json:"id"`
		Name               string `json:"name"`
		Provider           string `json:"provider"`
		ProviderRole       string `json:"providerRole"`
		Format             string `json:"format"`
		RepoType           string `json:"repoType"`
		RepoPolicy         string `json:"repoPolicy"`
		Exposed            bool   `json:"exposed"`
	}

	Client struct {
		baseURL    string // http://localhost:8081/nexus
		username   string
		password   string
		httpClient *http.Client
	}
)

func init() {
}

func NewClient(baseURL, username, password string) *Client {
	client := &http.Client{}
	return &Client{baseURL, username, password, client}
}

func (client *Client) RepositoryExists(repositoryID string) (bool, error) {
	req, err := http.NewRequest("GET", client.baseURL+"/service/local/repositories/"+repositoryID, nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return false, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return false, fmt.Errorf("Client.RepositoryExists(): unexpected response status: %d\n", resp.StatusCode)
	}

	return resp.StatusCode == 200, nil
}

func (client *Client) CreateRepository(repositoryID string) error {
	repo := createrepo{
		Data: Data{
			Id:                 repositoryID,
			Name:               repositoryID,
			Provider:           "maven2",
			RepoType:           "hosted",
			RepoPolicy:         "SNAPSHOT",
			ProviderRole:       "org.sonatype.nexus.proxy.repository.Repository",
			ContentResourceURI: client.baseURL + "/content/repositories/" + repositoryID,
			Format:             "maven2",
			Exposed:            true,
		}}

	data, err := json.Marshal(&repo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", client.baseURL+"/service/local/repositories", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("Client.CreateRepository(): unexpected response status: %d (%s)\n", resp.StatusCode, string(body))
	}

	return nil
}

func (client *Client) DeleteRepository(repositoryID string) error {
	req, err := http.NewRequest("DELETE", client.baseURL+"/service/local/repositories/"+repositoryID, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		return fmt.Errorf("Client.DeleteRepository(): unexpected response status: %d\n", resp.StatusCode)
	}

	return nil
}
