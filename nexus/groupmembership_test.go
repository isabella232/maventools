package nexus

import "testing"

func TestGroupMembership(t *testing.T) {
	ra := []repository{
		{ID: "foo", Name: "foo", ResourceURI: "blah"},
		{ID: "bar", Name: "bar", ResourceURI: "blah"},
	}
	group := RepoGroup{Data: RepositoryGroupData{Repositories: ra}}

	present := repoIsInGroup("foo", group)
	if !present {
		t.Fatalf("Wanted true but got false\n")
	}

	absent := repoIsNotInGroup("baz", group)
	if !absent {
		t.Fatalf("Wanted true but got false\n")
	}

}
