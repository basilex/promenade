package cache

import (
	"context"
	"sync"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/account/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// AccountCache provides in-memory caching for Chart of Accounts
type AccountCache struct {
	repo          repository.IAccountRepository
	mu            sync.RWMutex
	byID          map[uuidv7.UUID]*cacheEntry
	byCode        map[string]*cacheEntry
	byOrg         map[uuidv7.UUID][]*aggregate.Account
	ttl           time.Duration
	lastCacheTime map[uuidv7.UUID]time.Time
}

type cacheEntry struct {
	account   *aggregate.Account
	expiresAt time.Time
}

// NewAccountCache creates a new account cache with 5-minute TTL
func NewAccountCache(repo repository.IAccountRepository) *AccountCache {
	return &AccountCache{
		repo:          repo,
		byID:          make(map[uuidv7.UUID]*cacheEntry),
		byCode:        make(map[string]*cacheEntry),
		byOrg:         make(map[uuidv7.UUID][]*aggregate.Account),
		ttl:           5 * time.Minute,
		lastCacheTime: make(map[uuidv7.UUID]time.Time),
	}
}

// GetByID retrieves an account by ID from cache or repository
func (c *AccountCache) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
	c.mu.RLock()
	if entry, ok := c.byID[id]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.account, nil
	}
	c.mu.RUnlock()

	// Fetch from repository
	account, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	c.byID[id] = &cacheEntry{
		account:   account,
		expiresAt: time.Now().Add(c.ttl),
	}
	codeKey := c.codeKey(account.OrganizationID, account.Code)
	c.byCode[codeKey] = c.byID[id]
	c.mu.Unlock()

	return account, nil
}

// GetByCode retrieves an account by code from cache or repository
func (c *AccountCache) GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.Account, error) {
	codeKey := c.codeKey(organizationID, code)

	c.mu.RLock()
	if entry, ok := c.byCode[codeKey]; ok && !c.isExpired(entry) {
		c.mu.RUnlock()
		return entry.account, nil
	}
	c.mu.RUnlock()

	// Fetch from repository
	account, err := c.repo.GetByCode(ctx, organizationID, code)
	if err != nil {
		return nil, err
	}

	// Cache it
	c.mu.Lock()
	entry := &cacheEntry{
		account:   account,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.byID[account.ID] = entry
	c.byCode[codeKey] = entry
	c.mu.Unlock()

	return account, nil
}

// GetAllActive retrieves all active accounts for an organization
func (c *AccountCache) GetAllActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Account, error) {
	c.mu.RLock()
	if accounts, ok := c.byOrg[organizationID]; ok {
		if lastCache, exists := c.lastCacheTime[organizationID]; exists && time.Since(lastCache) < c.ttl {
			c.mu.RUnlock()
			return accounts, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from repository
	accounts, err := c.repo.GetAllActive(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	// Cache all accounts
	c.mu.Lock()
	c.byOrg[organizationID] = accounts
	c.lastCacheTime[organizationID] = time.Now()

	for _, acc := range accounts {
		entry := &cacheEntry{
			account:   acc,
			expiresAt: time.Now().Add(c.ttl),
		}
		c.byID[acc.ID] = entry
		codeKey := c.codeKey(acc.OrganizationID, acc.Code)
		c.byCode[codeKey] = entry
	}
	c.mu.Unlock()

	return accounts, nil
}

// Invalidate clears cache for an organization
func (c *AccountCache) Invalidate(organizationID uuidv7.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove organization cache
	delete(c.byOrg, organizationID)
	delete(c.lastCacheTime, organizationID)

	// Remove individual account caches for this organization
	for id, entry := range c.byID {
		if entry.account.OrganizationID == organizationID {
			delete(c.byID, id)
		}
	}

	// Remove code-based caches
	for key, entry := range c.byCode {
		if entry.account.OrganizationID == organizationID {
			delete(c.byCode, key)
		}
	}
}

// WarmCache preloads the cache with all active accounts for an organization
func (c *AccountCache) WarmCache(ctx context.Context, organizationID uuidv7.UUID) error {
	_, err := c.GetAllActive(ctx, organizationID)
	return err
}

// Stats returns cache statistics
func (c *AccountCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalAccounts := len(c.byID)
	return CacheStats{
		TTL:           c.ttl,
		TotalAccounts: totalAccounts,
		Organizations: len(c.byOrg),
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	TTL           time.Duration
	TotalAccounts int
	Organizations int
}

func (c *AccountCache) isExpired(entry *cacheEntry) bool {
	return time.Now().After(entry.expiresAt)
}

func (c *AccountCache) codeKey(organizationID uuidv7.UUID, code string) string {
	return organizationID.String() + ":" + code
}
