package store

import "github.com/roemer/test-tamer/internal/repositories"

// Repositories is a collection of all repository interfaces.
type Repositories struct {
	Project repositories.ProjectRepository
}

// NewDbRepositories creates a new Repositories instance with the provided DBTX.
func NewDbRepositories(db repositories.DBTX) *Repositories {
	return &Repositories{
		Project: repositories.NewProjectRepository(db),
	}
}
