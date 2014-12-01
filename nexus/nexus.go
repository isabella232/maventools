package nexus

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	baseURL    string // http://localhost:8081/nexus
	username   string
	password   string
	httpClient *http.Client
}

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
