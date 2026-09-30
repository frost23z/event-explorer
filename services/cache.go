package services

import (
	"sync"
	"time"

	"event-explorer/models"

	"github.com/beego/beego/v2/core/logs"
)

const cacheTTL = 5 * time.Minute

type cacheEntry struct {
	events    []models.Event
	expiresAt time.Time
}

// eventCache is one map shared by every request (and by the Music and Sports
// goroutines), so each access holds the mutex. Callers must not modify the
// returned slices.
type eventCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	now     func() time.Time // replaced in tests
}

func newEventCache() *eventCache {
	return &eventCache{entries: map[string]cacheEntry{}, now: time.Now}
}

// cacheKey is city, country and category: the same values Ticketmaster is asked for.
func cacheKey(params models.EventSearchParams) string {
	return params.City + "|" + params.CountryCode + "|" + params.Category
}

func (c *eventCache) get(key string) ([]models.Event, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if !c.now().Before(entry.expiresAt) {
		delete(c.entries, key)
		return nil, false
	}

	logs.Info("cache hit: %s", key)

	return entry.events, true
}

func (c *eventCache) set(key string, events []models.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{events: events, expiresAt: c.now().Add(cacheTTL)}
}
