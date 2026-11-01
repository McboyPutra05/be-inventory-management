//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intern/inventory-api/internal/config"
	"github.com/intern/inventory-api/internal/database"
	"github.com/intern/inventory-api/internal/domain"
	"github.com/intern/inventory-api/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestConcurrency_AtomicStockDeduction(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Skipping integration test: configuration not loaded")
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := database.RunMigrations(sqlDB); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	itemRepo := repository.NewItemRepository(db)

	ctx := context.Background()
	initialStock := 10
	totalGoroutines := 50

	item := &domain.Item{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Stress Item %s", uuid.NewString()[:8]),
		Description: "Concurrency test item",
		Stock:       initialStock,
		Price:       15000,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err = itemRepo.Create(ctx, item)
	assert.NoError(t, err)
	defer itemRepo.Delete(ctx, item.ID)

	var wg sync.WaitGroup
	var successCount int64
	var insufficientCount int64

	startSignal := make(chan struct{})

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal

			_, adjErr := itemRepo.AdjustStock(ctx, item.ID, -1)
			if adjErr == nil {
				atomic.AddInt64(&successCount, 1)
			} else if adjErr == domain.ErrInsufficientStock {
				atomic.AddInt64(&insufficientCount, 1)
			}
		}()
	}

	close(startSignal)
	wg.Wait()

	assert.Equal(t, int64(10), successCount, "Exactly 10 requests should succeed")
	assert.Equal(t, int64(40), insufficientCount, "Exactly 40 requests should return ErrInsufficientStock")

	finalItem, err := itemRepo.FindByID(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, finalItem.Stock, "Final stock must be exactly 0, never negative")
}

