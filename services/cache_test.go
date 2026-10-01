package services

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"event-explorer/models"
)

func eventsWithIDs(ids ...string) []models.Event {
	events := make([]models.Event, 0, len(ids))

	for _, id := range ids {
		events = append(events, models.Event{ID: id, Name: "Event " + id})
	}

	return events
}

// newClockedCache returns a cache whose clock the test moves by hand.
func newClockedCache() (*eventCache, *time.Time) {
	now := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)

	cache := newEventCache()
	cache.now = func() time.Time { return now }

	return cache, &now
}

func TestCacheMiss(t *testing.T) {
	cache, _ := newClockedCache()

	if events, ok := cache.get("missing"); ok || events != nil {
		t.Fatalf("got %v, %v; want a miss", events, ok)
	}
}

func TestCacheHit(t *testing.T) {
	cache, _ := newClockedCache()
	cache.set("toronto|CA|music", eventsWithIDs("e1", "e2"))

	got, ok := cache.get("toronto|CA|music")
	if !ok || len(got) != 2 || got[0].ID != "e1" || got[1].ID != "e2" {
		t.Fatalf("got %v, %v; want the stored events", got, ok)
	}
}

func TestCacheHitForEmptyList(t *testing.T) {
	cache, _ := newClockedCache()
	cache.set("k", []models.Event{})

	got, ok := cache.get("k")
	if !ok || len(got) != 0 {
		t.Fatalf("got %v, %v; an empty successful list is still a hit", got, ok)
	}
}

func TestCacheExpiresAfterFiveMinutes(t *testing.T) {
	if cacheTTL != 5*time.Minute {
		t.Fatalf("cacheTTL = %v, the requirement is five minutes", cacheTTL)
	}

	cache, now := newClockedCache()
	cache.set("k", eventsWithIDs("e1"))

	*now = now.Add(cacheTTL - time.Nanosecond)

	if _, ok := cache.get("k"); !ok {
		t.Fatal("entry should still be fresh just before the TTL")
	}

	*now = now.Add(time.Nanosecond)

	if _, ok := cache.get("k"); ok {
		t.Fatal("entry should be expired exactly at the TTL")
	}
}

func TestCacheRemovesExpiredEntry(t *testing.T) {
	cache, now := newClockedCache()
	cache.set("k", eventsWithIDs("e1"))

	*now = now.Add(cacheTTL + time.Second)

	if _, ok := cache.get("k"); ok {
		t.Fatal("entry should be expired")
	}

	if len(cache.entries) != 0 {
		t.Fatalf("expired entry was not removed: %v", cache.entries)
	}
}

func TestCacheSetRefreshesExpiredEntry(t *testing.T) {
	cache, now := newClockedCache()
	cache.set("k", eventsWithIDs("old"))

	*now = now.Add(cacheTTL + time.Second)
	cache.set("k", eventsWithIDs("new"))

	// The refreshed entry gets a full new TTL.
	*now = now.Add(cacheTTL - time.Second)

	got, ok := cache.get("k")
	if !ok || len(got) != 1 || got[0].ID != "new" {
		t.Fatalf("got %v, %v; want the refreshed entry", got, ok)
	}
}

func TestCacheKeyUsesCityCountryAndCategory(t *testing.T) {
	base := models.EventSearchParams{City: "toronto", CountryCode: "CA", Category: models.CategoryMusic}

	first, second := cacheKey(base), cacheKey(base)
	if first != second {
		t.Fatal("the same search must give the same key")
	}

	variants := map[string]models.EventSearchParams{
		"city":     {City: "ottawa", CountryCode: "CA", Category: models.CategoryMusic},
		"country":  {City: "toronto", CountryCode: "US", Category: models.CategoryMusic},
		"category": {City: "toronto", CountryCode: "CA", Category: models.CategorySports},
	}

	for name, params := range variants {
		if cacheKey(params) == cacheKey(base) {
			t.Errorf("changing the %s must change the key", name)
		}
	}
}

func TestCacheKeysDoNotShareEntries(t *testing.T) {
	cache, _ := newClockedCache()

	music := models.EventSearchParams{City: "toronto", CountryCode: "CA", Category: models.CategoryMusic}
	sports := models.EventSearchParams{City: "toronto", CountryCode: "CA", Category: models.CategorySports}

	cache.set(cacheKey(music), eventsWithIDs("m1"))
	cache.set(cacheKey(sports), eventsWithIDs("s1"))

	if got, _ := cache.get(cacheKey(music)); len(got) != 1 || got[0].ID != "m1" {
		t.Errorf("music = %v", got)
	}

	if got, _ := cache.get(cacheKey(sports)); len(got) != 1 || got[0].ID != "s1" {
		t.Errorf("sports = %v", got)
	}
}

// Run with -race: every goroutine reads and writes the same keys.
func TestCacheConcurrentAccess(t *testing.T) {
	cache := newEventCache()

	var wg sync.WaitGroup

	for worker := range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := range 200 {
				key := "key-" + strconv.Itoa(i%5)

				cache.set(key, eventsWithIDs("w"+strconv.Itoa(worker)))
				cache.get(key)
			}
		}()
	}

	wg.Wait()

	if len(cache.entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(cache.entries))
	}
}
func TestCacheClearRemovesEveryEntry(t *testing.T) {
	cache, _ := newClockedCache()
	cache.set("a", eventsWithIDs("1"))
	cache.set("b", eventsWithIDs("2"))

	if got := cache.clear(); got != 2 {
		t.Fatalf("cleared %d entries, want 2", got)
	}

	for _, key := range []string{"a", "b"} {
		if _, ok := cache.get(key); ok {
			t.Errorf("%s is still cached after clear", key)
		}
	}

	if got := cache.clear(); got != 0 {
		t.Errorf("clearing an empty cache removed %d entries, want 0", got)
	}
}
