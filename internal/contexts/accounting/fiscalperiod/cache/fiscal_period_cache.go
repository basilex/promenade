package cache

import (
	"context"
	"sync"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// FiscalPeriodCache provides in-memory caching for fiscal periods
type FiscalPeriodCache struct {
	repo          repository.IFiscalPeriodRepository
	mu            sync.RWMutex
	byID          map[uuidv7.UUID]*periodCacheEntry
	byOrgDate     map[string]*periodCacheEntry
	byOrg         map[uuidv7.UUID][]*aggregate.FiscalPeriod
	ttl           time.Duration
	lastCacheTime map[uuidv7.UUID]time.Time
}

type periodCacheEntry struct {
	period    *aggregate.FiscalPeriod
	expiresAt time.Time
}

// NewFiscalPeriodCache creates a new fiscal period cache with 10-minute TTL
func NewFiscalPeriodCache(repo repository.IFiscalPeriodRepository) *FiscalPeriodCache {
	return &FiscalPeriodCache{
		repo:          repo,
		byID:          make(map[uuidv7.UUID]*periodCacheEntry),
		byOrgDate:     make(map[string]*periodCacheEntry),
		byOrg:         make(map[uuidv7.UUID][]*aggregate.FiscalPeriod),
		ttl:           10 * time.Minute,
		lastCacheTime: make(map[uuidv7.UUID]time.Time),
	}
}

// GetByID retrieves a fiscal period by ID from cache or repository
func (c *FiscalPeriodCache) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	c.mu.RLock()
	if entry, ok := c.byID[id]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.period, nil
	}
	c.mu.RUnlock()

	// Fetch from repository
	period, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	c.byID[id] = &periodCacheEntry{
		period:    period,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return period, nil
}

// GetByOrganizationAndDate retrieves a fiscal period containing the given date
func (c *FiscalPeriodCache) GetByOrganizationAndDate(ctx context.Context, organizationID uuidv7.UUID, date time.Time) (*aggregate.FiscalPeriod, error) {
	dateKey := c.dateKey(organizationID, date)

	c.mu.RLock()
	if entry, ok := c.byOrgDate[dateKey]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.period, nil
	}
	c.mu.RUnlock()

	// Check if we have periods cached for this organization
	c.mu.RLock()
	if periods, ok := c.byOrg[organizationID]; ok {
		if lastCache, exists := c.lastCacheTime[organizationID]; exists && time.Since(lastCache) < c.ttl {
			// Search in cached periods
			for _, period := range periods {
				if (date.Equal(period.StartDate) || date.After(period.StartDate)) &&
					(date.Equal(period.EndDate) || date.Before(period.EndDate)) {
					c.mu.RUnlock()
					return period, nil
				}
			}
		}
	}
	c.mu.RUnlock()

	// Fetch from repository - convert date to string format YYYY-MM-DD
	dateStr := date.Format("2006-01-02")
	period, err := c.repo.GetByOrganizationAndDate(ctx, organizationID, dateStr)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	entry := &periodCacheEntry{
		period:    period,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.byID[period.ID] = entry
	c.byOrgDate[dateKey] = entry
	c.mu.Unlock()

	return period, nil
}

// ListOpen retrieves all open periods for an organization (for posting validation)
func (c *FiscalPeriodCache) ListOpen(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
	// Check if we have fresh cached data
	c.mu.RLock()
	if periods, ok := c.byOrg[organizationID]; ok {
		if lastCache, exists := c.lastCacheTime[organizationID]; exists && time.Since(lastCache) < c.ttl {
			// Filter open periods
			var openPeriods []*aggregate.FiscalPeriod
			for _, p := range periods {
				if p.Status == aggregate.PeriodStatusOpen {
					openPeriods = append(openPeriods, p)
				}
			}
			c.mu.RUnlock()
			return openPeriods, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from repository - get all periods without pagination
	periods, err := c.repo.ListByOrganization(ctx, organizationID, 0, 1000)
	if err != nil {
		return nil, err
	}

	// Cache all periods
	c.mu.Lock()
	c.byOrg[organizationID] = periods
	c.lastCacheTime[organizationID] = time.Now()

	for _, p := range periods {
		entry := &periodCacheEntry{
			period:    p,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.byID[p.ID] = entry
	}
	c.mu.Unlock()

	// Filter open periods
	var openPeriods []*aggregate.FiscalPeriod
	for _, p := range periods {
		if p.Status == aggregate.PeriodStatusOpen {
			openPeriods = append(openPeriods, p)
		}
	}

	return openPeriods, nil
}

// Invalidate clears cache for an organization
func (c *FiscalPeriodCache) Invalidate(organizationID uuidv7.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.byOrg, organizationID)
	delete(c.lastCacheTime, organizationID)

	for id, entry := range c.byID {
		if entry.period.OrganizationID == organizationID {
			delete(c.byID, id)
		}
	}

	for key, entry := range c.byOrgDate {
		if entry.period.OrganizationID == organizationID {
			delete(c.byOrgDate, key)
		}
	}
}

// WarmCache preloads the cache with all periods for an organization
func (c *FiscalPeriodCache) WarmCache(ctx context.Context, organizationID uuidv7.UUID) error {
	_, err := c.ListOpen(ctx, organizationID)
	return err
}

// Stats returns cache statistics
func (c *FiscalPeriodCache) Stats() PeriodCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalPeriods := len(c.byID)
	totalOpen := 0
	for _, entry := range c.byID {
		if entry.period.Status == aggregate.PeriodStatusOpen {
			totalOpen++
		}
	}

	return PeriodCacheStats{
		TTL:           c.ttl,
		OpenPeriods:   totalOpen,
		TotalPeriods:  totalPeriods,
		Organizations: len(c.byOrg),
	}
}

// PeriodCacheStats holds cache statistics
type PeriodCacheStats struct {
	TTL           time.Duration
	OpenPeriods   int
	TotalPeriods  int
	Organizations int
}

func (c *FiscalPeriodCache) isExpired(entry *periodCacheEntry) bool {
	return time.Now().After(entry.expiresAt)
}

func (c *FiscalPeriodCache) dateKey(organizationID uuidv7.UUID, date time.Time) string {
	return organizationID.String() + ":" + date.Format("2006-01-02")
}
