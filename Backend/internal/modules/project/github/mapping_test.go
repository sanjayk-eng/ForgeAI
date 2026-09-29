package github

import (
	"reflect"
	"testing"

	"ai-agent/internal/modules/auth/provider"
)

func TestRepositoryMappings(t *testing.T) {
	providerRepository := provider.GitHubRepository{
		ID: 42, Owner: "forge", Name: "agent", URL: "https://github.com/forge/agent",
		Private: true, DefaultBranch: "main", Branches: []string{"main", "develop"},
	}

	repository := mapRepository(providerRepository)
	if repository.ID != providerRepository.ID || repository.Owner != providerRepository.Owner ||
		repository.Name != providerRepository.Name || repository.URL != providerRepository.URL ||
		repository.Private != providerRepository.Private || repository.DefaultBranch != providerRepository.DefaultBranch ||
		!reflect.DeepEqual(repository.Branches, providerRepository.Branches) {
		t.Fatalf("mapRepository() = %+v, want fields copied from %+v", repository, providerRepository)
	}

	request := toConnectRepositoryRequest(repository)
	if request.GitHubRepositoryID != repository.ID || request.GitHubOwner != repository.Owner ||
		request.GitHubRepositoryName != repository.Name || request.RepositoryURL != repository.URL ||
		request.Private != repository.Private || request.DefaultBranch != repository.DefaultBranch {
		t.Fatalf("toConnectRepositoryRequest() = %+v, want fields copied from %+v", request, repository)
	}
}
