package maventools

import "net/http"

type RepositoryID string

type GroupID string

type Repository struct {
	ID   RepositoryID
	Name string
}

// Group models a repository group.
type Group struct {
	ID           GroupID
	Name         string
	Repositories []Repository
}

type Client struct {
	ClientOps
	BaseURL    string
	Username   string
	Password   string
	HttpClient *http.Client
}

// ClientOps defines the service methods on a Client. Integer return values are the underlying HTTP response codes.
type ClientOps interface {
	RepositoryExists(RepositoryID) (bool, error)
	CreateSnapshotRepository(RepositoryID) (int, error)
	DeleteRepository(RepositoryID) (int, error)
	AddRepositoryToGroup(RepositoryID, GroupID) (int, error)
	RemoveRepositoryFromGroup(RepositoryID, GroupID) (int, error)
}

/*
func NewNexusClient(baseURL, username, password string) nexus.Client {
	return nexus.Client{}
}
*/
