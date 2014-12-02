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
		Data CreateRepoData `json:"data"`
	}

	CreateRepoData struct {
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

	RepoGroup struct {
		Data RepositoryGroupData `json:"data"`
	}

	repository struct {
		Name        string `json:"name"`
		ID          string `json:"id"`
		ResourceURI string `json:resourceURI"`
	}

	RepositoryGroupData struct {
		ID                 string       `json:"id"`
		Provider           string       `json:"provider"`
		Name               string       `json:"name"`
		Repositories       []repository `json:"repositories"`
		Format             string       `json:"format"`
		RepoType           string       `json:"repoType"`
		Exposed            bool         `json:"exposed"`
		ContentResourceURI string       `json:"contentResourceURI"`
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
		Data: CreateRepoData{
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

// RepositoryGroup gets a repository group based on the given group ID.
func (client *Client) RepositoryGroup(groupID string) (RepoGroup, error) {
	req, err := http.NewRequest("GET", client.baseURL+"/service/local/repo_group/"+groupID, nil)
	if err != nil {
		return RepoGroup{}, err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return RepoGroup{}, err

	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RepoGroup{}, err
	}

	if resp.StatusCode != 200 {
		return RepoGroup{}, fmt.Errorf("Client.RepositoryGroup(): unexpected response status: %d\n", resp.StatusCode)
	}

	var repogroup RepoGroup
	if err := json.Unmarshal(data, &repogroup); err != nil {
		return RepoGroup{}, err
	}
	return repogroup, nil
}
