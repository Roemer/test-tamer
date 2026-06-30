package memory

import (
	"context"
	"sync"

	"github.com/roemer/test-tamer/internal/store"
)

type memoryClient struct {
	mu           sync.Mutex
	projectStore store.ProjectStore
}

func New() store.Client {
	return &memoryClient{
		projectStore: newProjectStore(),
	}
}

func (c *memoryClient) Transact(ctx context.Context, fn func(ctx context.Context, stores *store.Stores) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fn(ctx, c.Stores())
}

func (c *memoryClient) Stores() *store.Stores {
	return &store.Stores{
		Projects: c.projectStore,
	}
}
