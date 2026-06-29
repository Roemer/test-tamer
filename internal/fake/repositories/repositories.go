package repositories

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/entities"
	"github.com/roemer/test-tamer/internal/store"
)

// NewFakeRepositories creates a new Repositories instance with a provided project repository.
func NewFakeRepositories() *store.Repositories {
	projects := []entities.Project{}
	for i := range 350 {
		project := entities.Project{
			BaseEntity: entities.BaseEntity{
				ID:        int64(i + 1),
				PublicID:  uuid.Must(uuid.NewV7()),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: "Project " + strconv.Itoa(i+1),
		}
		projects = append(projects, project)
	}
	return &store.Repositories{
		Project: NewFakeProjectRepository(projects),
	}
}

func addRandomDelay() {
	// Simulate a random delay between 100ms and 800ms
	time.Sleep(time.Duration(100+time.Now().UnixNano()%700) * time.Millisecond)
}
