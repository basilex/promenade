package cache

import (
	"context"
	"sync"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TaxCodeCache provides in-memory caching for tax codes
type TaxCodeCache struct {
	repo          repository.ITaxCodeRepository
	mu            sync.RWMutex
	byID          map[uuidv7.UUID]*taxCacheEntry
	byCode        map[string]*taxCacheEntry
	byOrg         map[uuidv7.UUID][]*aggregate.TaxCode
	ttl           time.Duration
	lastCacheTime map[uuidv7.UUID]time.Time
}

type taxCacheEntry struct {
	taxCode   *aggregate.TaxCode
	expiresAt time.Time
}

// NewTaxCodeCache creates a new tax code cache with 15-minute TTL (longer for static data)
func NewTaxCodeCache(repo repository.ITaxCodeRepository) *TaxCodeCache {
	return &TaxCodeCache{
		repo:          repo,
		byID:          make(map[uuidv7.UUID]*taxCacheEntry),
		byCode:        make(map[string]*taxCacheEntry),
		byOrg:         make(map[uuidv7.UUID][]*aggregate.TaxCode),
		ttl:           15 * time.Minute,
		lastCacheTime: make(map[uuidv7.UUID]time.Time),
	}
}

// GetByID retrieves a tax code by ID from cache or repository
func (c *TaxCodeCache) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
	c.mu.RLock()
	if entry, ok := c.byID[id]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.taxCode, nil
	}
	c.mu.RUnlock()

	// Fetch from repository
	taxCode, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	c.byID[id] = &taxCacheEntry{
		taxCode:   taxCode,
		expiresAt: time.Now().Add(c.ttl),
	}
	codeKey := c.codeKey(taxCode.OrganizationID, taxCode.Code)
	c.byCode[codeKey] = c.byID[id]
	c.mu.Unlock()

	return taxCode, nil
}

// GetByCode retrieves a tax code by code from cache or repository
func (c *TaxCodeCache) GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
	codeKey := c.codeKey(organizationID, code)

	c.mu.RLock()
	if entry, ok := c.byCode[codeKey]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.taxCode, nil
	}
	c.mu.RUnlock()

	// Fetch from repository
	taxCode, err := c.repo.GetByCode(ctx, organizationID, code)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	entry := &taxCacheEntry{
		taxCode:   taxCode,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.byID[taxCode.ID] = entry
	c.byCode[codeKey] = entry
	c.mu.Unlock()

	return taxCode, nil
}

// GetAllActive retrieves all active tax codes for an organization
func (c *TaxCodeCache) GetAllActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.TaxCode, error) {
	c.mu.RLock()
	if taxCodes, ok := c.byOrg[organizationID]; ok {
		if lastCache, exists := c.lastCacheTime[organizationID]; exists && time.Since(lastCache) < c.ttl {
			c.mu.RUnlock()
			return taxCodes, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from repository - get all tax codes without pagination
	taxCodes, err := c.repo.ListByOrganization(ctx, organizationID, 0, 1000)
	if err != nil {
		return nil, err
	}

	// Filter active only
	var activeTaxCodes []*aggregate.TaxCode
	for _, tc := range taxCodes {
		if tc.IsActive {
			activeTaxCodes = append(activeTaxCodes, tc)
		}
	}
	taxCodes = activeTaxCodes

	// Cache all tax codes
	c.mu.Lock()
	c.byOrg[organizationID] = taxCodes
	c.lastCacheTime[organizationID] = time.Now()

	for _, tc := range taxCodes {
		entry := &taxCacheEntry{
			taxCode:   tc,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.byID[tc.ID] = entry
		codeKey := c.codeKey(tc.OrganizationID, tc.Code)
		c.byCode[codeKey] = entry
	}
	c.mu.Unlock()

	return taxCodes, nil
}

// Invalidate clears cache for an organization
func (c *TaxCodeCache) Invalidate(organizationID uuidv7.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.byOrg, organizationID)
	delete(c.lastCacheTime, organizationID)

	for id, entry := range c.byID {
		if entry.taxCode.OrganizationID == organizationID {
			delete(c.byID, id)
		}
	}

	for key, entry := range c.byCode {
		if entry.taxCode.OrganizationID == organizationID {
			delete(c.byCode, key)
		}
	}
}

// WarmCache preloads the cache with all active tax codes for an organization
func (c *TaxCodeCache) WarmCache(ctx context.Context, organizationID uuidv7.UUID) error {
	_, err := c.GetAllActive(ctx, organizationID)
	return err
}

func (c *TaxCodeCache) isExpired(entry *taxCacheEntry) bool {
	return time.Now().After(entry.expiresAt)
}

func (c *TaxCodeCache) codeKey(organizationID uuidv7.UUID, code string) string {
	return organizationID.String() + ":" + code
}
