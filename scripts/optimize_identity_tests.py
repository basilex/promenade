#!/usr/bin/env python3
"""
Script to generate optimized integration test files for Identity context.
Generates CRUD + Queries pattern tests for each aggregate.
"""

import os

def generate_profile_test():
    content = """package profile_test

import (
\t"context"
\t"fmt"
\t"testing"

\t"github.com/jmoiron/sqlx"
\t"github.com/stretchr/testify/assert"
\t"github.com/stretchr/testify/require"

\t"github.com/basilex/promenade/internal/contexts/identity/profile"
\t"github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
\t"github.com/basilex/promenade/pkg/uuidv7"
\t"github.com/basilex/promenade/test/integration"
)

func TestProfileRepository_CRUD(t *testing.T) {
\tif testing.Short() {
\t\tt.Skip("Skipping integration test in short mode")
\t}
\ttestDB := integration.SetupTestDB(t)
\ttestDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
\t\trepo := postgres.NewProfileRepository(tx)
\t\tuserID := uuidv7.New()

\t\t// Create user
\t\t_, err := tx.Exec(`INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
\t\t\tuserID, fmt.Sprintf("user_%s@test.com", userID), "hash", "active")
\t\trequire.NoError(t, err)

\t\t// Create
\t\tp, err := profile.NewProfile(userID, "Test User")
\t\trequire.NoError(t, err)
\t\trequire.NoError(t, repo.Create(ctx, p))

\t\t// Read by ID
\t\tfound, err := repo.GetByID(ctx, p.ID)
\t\trequire.NoError(t, err)
\t\tassert.Equal(t, "Test User", found.DisplayName)
\t\tassert.Equal(t, userID, found.UserID)

\t\t// Read by UserID
\t\tfoundByUser, err := repo.GetByUserID(ctx, userID)
\t\trequire.NoError(t, err)
\t\tassert.Equal(t, p.ID, foundByUser.ID)

\t\t// Update
\t\tfound.UpdateDisplayName("Updated User")
\t\tfound.UpdateBio("New bio")
\t\trequire.NoError(t, repo.Update(ctx, found))
\t\tupdated, _ := repo.GetByID(ctx, p.ID)
\t\tassert.Equal(t, "Updated User", updated.DisplayName)
\t\tassert.Equal(t, "New bio", updated.Bio)

\t\t// Delete
\t\trequire.NoError(t, repo.Delete(ctx, p.ID))
\t\t_, err = repo.GetByID(ctx, p.ID)
\t\tassert.Error(t, err)
\t})
}

func TestProfileRepository_Queries(t *testing.T) {
\tif testing.Short() {
\t\tt.Skip("Skipping integration test in short mode")
\t}
\ttestDB := integration.SetupTestDB(t)
\ttestDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
\t\trepo := postgres.NewProfileRepository(tx)

\t\t// Create 3 profiles (2 public, 1 private)
\t\tfor i := 0; i < 3; i++ {
\t\t\tuserID := uuidv7.New()
\t\t\t_, err := tx.Exec(`INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
\t\t\t\tuserID, fmt.Sprintf("user%d@test.com", i), "hash", "active")
\t\t\trequire.NoError(t, err)

\t\t\tp, _ := profile.NewProfile(userID, fmt.Sprintf("User %d", i))
\t\t\tif i < 2 {
\t\t\t\tp.SetPublic() // Make first 2 public
\t\t\t}
\t\t\trequire.NoError(t, repo.Create(ctx, p))
\t\t}

\t\t// List public profiles
\t\tpublicProfiles, err := repo.ListPublicProfiles(ctx, 10, 0)
\t\trequire.NoError(t, err)
\t\tassert.GreaterOrEqual(t, len(publicProfiles), 2)
\t\tfor _, p := range publicProfiles {
\t\t\tassert.True(t, p.IsPublic)
\t\t}
\t})
}
"""
    return content

def main():
    # Generate Profile test
    profile_test = generate_profile_test()
    output_path = "test/integration/contexts/identity/profile/repository_test.go"
    
    with open(output_path, 'w') as f:
        f.write(profile_test)
    
    print(f"✓ Generated {output_path}")
    print(f"  Lines: {len(profile_test.splitlines())}")

if __name__ == "__main__":
    main()
