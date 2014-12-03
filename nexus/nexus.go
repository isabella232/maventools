package nexus

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xoom/maventools"
)

type (
	createrepo struct {
		XMLName xml.Name       `xml:"repository"`
		Data    CreateRepoData `xml:"data"`
	}

	CreateRepoData struct {
		XMLName            xml.Name                `xml:"data"`
		ContentResourceURI string                  `xml:"contentResourceURI"`
		Id                 maventools.RepositoryID `xml:"id"`
		Name               string                  `xml:"name"`
		Provider           string                  `xml:"provider"`
		ProviderRole       string                  `xml:"providerRole"`
		Format             string                  `xml:"format"`
		RepoType           string                  `xml:"repoType"`
		RepoPolicy         string                  `xml:"repoPolicy"`
		Exposed            bool                    `xml:"exposed"`
	}

	// The type retrieved or put to read or mutate a repository group.
	RepoGroup struct {
		Data RepositoryGroupData `json:"data"`
	}

	// The payload of a repository group read or mutation.
	RepositoryGroupData struct {
		ID                 maventools.GroupID `json:"id"`
		Provider           string             `json:"provider"`
		Name               string             `json:"name"`
		Repositories       []repository       `json:"repositories"`
		Format             string             `json:"format"`
		RepoType           string             `json:"repoType"`
		Exposed            bool               `json:"exposed"`
		ContentResourceURI string             `json:"contentResourceURI"`
	}

	repository struct {
		Name        string                  `json:"name"`
		ID          maventools.RepositoryID `json:"id"`
		ResourceURI string                  `json:resourceURI"`
	}

	// A Nexus client
	Client struct {
		baseURL    string // http://localhost:8081/nexus
		username   string
		password   string
		httpClient *http.Client
	}
)

// NewClient creates a new Nexus client on which subsequent service methods are called.  The baseURL typically takes
// the form http://host:port/nexus.  username and password are the credentials of an admin user capable of creating and mutating data
// within Nexus.
func NewClient(baseURL, username, password string) *Client {
	return &Client{baseURL, username, password, &http.Client{}}
}

// RepositoryExists checks whether a given repository specified by repositoryID exists.
func (client *Client) RepositoryExists(repositoryID maventools.RepositoryID) (bool, error) {
	req, err := http.NewRequest("GET", client.baseURL+"/service/local/repositories/"+string(repositoryID), nil)
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

// CreateSnapshotRepository creates a new hosted Maven2 SNAPSHOT repository with the given repositoryID.  The repository name
// will be the same as the repositoryID.  When error is nil, the integer return value is the underlying HTTP response code.
func (client *Client) CreateSnapshotRepository(repositoryID maventools.RepositoryID) (int, error) {
	repo := createrepo{
		Data: CreateRepoData{
			Id:                 repositoryID,
			Name:               string(repositoryID),
			Provider:           "maven2",
			RepoType:           "hosted",
			RepoPolicy:         "SNAPSHOT",
			ProviderRole:       "org.sonatype.nexus.proxy.repository.Repository",
			ContentResourceURI: client.baseURL + "/content/repositories/" + string(repositoryID),
			Format:             "maven2",
			Exposed:            true,
		}}

	data, err := xml.Marshal(&repo)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", client.baseURL+"/service/local/repositories", bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Add("Content-type", "application/xml")
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 201 {
		return resp.StatusCode, fmt.Errorf("Client.CreateSnapshotRepository(): unexpected response status: %d (%s)\n", resp.StatusCode, string(body))
	}

	return resp.StatusCode, nil
}

// DeleteRepository deletes the repository with the given repositoryID.
func (client *Client) DeleteRepository(repositoryID maventools.RepositoryID) error {
	req, err := http.NewRequest("DELETE", client.baseURL+"/service/local/repositories/"+string(repositoryID), nil)
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

// RepositoryGroup gets a repository group specified by groupID.
func (client *Client) RepositoryGroup(groupID maventools.GroupID) (RepoGroup, error) {
	req, err := http.NewRequest("GET", client.baseURL+"/service/local/repo_groups/"+string(groupID), nil)
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

// Add RepositoryToGroup adds the given repository specified by repositoryID to the repository group specified by groupID.
func (client *Client) AddRepositoryToGroup(repositoryID maventools.RepositoryID, groupID maventools.GroupID) error {
	repogroup, err := client.RepositoryGroup(groupID)
	if err != nil {
		return err
	}

	if repoIsInGroup(repositoryID, repogroup) {
		return nil
	}

	repo := repository{ID: repositoryID, Name: string(repositoryID), ResourceURI: client.baseURL + "/service/local/repo_groups/" + string(groupID) + "/" + string(repositoryID)}
	repogroup.Data.Repositories = append(repogroup.Data.Repositories, repo)

	data, err := json.Marshal(&repogroup)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", client.baseURL+"/service/local/repo_groups/"+string(groupID), bytes.NewBuffer(data))
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

	if resp.StatusCode != 200 {
		return fmt.Errorf("Client.AddRepositoryToGroup(): unexpected response status: %d (%s)\n", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteRepositoryFromGroup removes the given repository specified by repositoryID from the repository group specified by groupID.
func (client *Client) DeleteRepositoryFromGroup(repositoryID maventools.RepositoryID, groupID maventools.GroupID) error {
	repogroup, err := client.RepositoryGroup(groupID)
	if err != nil {
		return err
	}

	if repoIsNotInGroup(repositoryID, repogroup) {
		return nil
	}

	removeRepo(repositoryID, &repogroup)

	data, err := json.Marshal(&repogroup)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", client.baseURL+"/service/local/repo_groups/"+string(groupID), bytes.NewBuffer(data))
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

	if resp.StatusCode != 200 {
		return fmt.Errorf("Client.AddRepositoryToGroup(): unexpected response status: %d (%s)\n", resp.StatusCode, string(body))
	}

	return nil
}

func repoIsInGroup(repositoryID maventools.RepositoryID, group RepoGroup) bool {
	for _, repo := range group.Data.Repositories {
		if repo.ID == repositoryID {
			return true
		}
	}
	return false
}

func repoIsNotInGroup(repositoryID maventools.RepositoryID, group RepoGroup) bool {
	for _, repo := range group.Data.Repositories {
		if repo.ID == repositoryID {
			return false
		}
	}
	return true
}

func removeRepo(repositoryID maventools.RepositoryID, group *RepoGroup) {
	ra := make([]repository, 0)
	for _, repo := range group.Data.Repositories {
		if repo.ID != repositoryID {
			ra = append(ra, repo)
		}
	}
	group.Data.Repositories = ra
}
