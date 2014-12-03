package maventools

import "net/http"

type RepositoryID string

type GroupID string

type ClientConfig struct {
	Client
	BaseURL    string
	Username   string
	Password   string
	HttpClient *http.Client
}

// ClientOps defines the service methods on a Client. Integer return values are the underlying HTTP response codes.
type Client interface {
	RepositoryExists(RepositoryID) (bool, error)
	CreateSnapshotRepository(RepositoryID) (int, error)
	DeleteRepository(RepositoryID) (int, error)
	AddRepositoryToGroup(RepositoryID, GroupID) (int, error)
	RemoveRepositoryFromGroup(RepositoryID, GroupID) (int, error)
}
