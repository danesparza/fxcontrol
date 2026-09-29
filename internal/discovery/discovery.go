// Package discovery browses FX advertisements and maintains an in-memory snapshot.
package discovery

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/danesparza/fxcontrol/internal/mapper"
	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"
)

const (
	ServiceType     = "_fx._tcp"
	Domain          = "local."
	ScanWindow      = 5 * time.Second
	RefreshInterval = 30 * time.Second
)

// Scanner performs a bounded network scan. Cancellation must stop the scan.
type Scanner interface {
	Scan(context.Context) ([]model.Service, error)
}

// NetworkScanner browses the same mDNS service type advertised by the FX services.
type NetworkScanner struct{}

func (NetworkScanner) Scan(ctx context.Context) ([]model.Service, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("create discovery resolver: %w", err)
	}
	scanCtx, cancel := context.WithTimeout(ctx, ScanWindow)
	defer cancel()
	entries := make(chan *zeroconf.ServiceEntry, 32)
	if err := resolver.Browse(scanCtx, ServiceType, Domain, entries); err != nil {
		return nil, fmt.Errorf("browse FX services: %w", err)
	}
	services := make([]model.Service, 0)
	// Browse closes entries after cancellation and releases its multicast sockets.
	for entry := range entries {
		if service, ok := mapper.DiscoveryService(entry); ok {
			services = append(services, service)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if scanCtx.Err() == nil {
		return nil, fmt.Errorf("discovery browser stopped before scan completed")
	}
	return services, nil
}

// Cache allows HTTP readers to access the last completed scan without waiting.
// Its zero value is ready for use.
type Cache struct {
	mu       sync.RWMutex
	services []model.Service
}

func clone(services []model.Service) []model.Service {
	result := make([]model.Service, len(services))
	copy(result, services)
	for i := range result {
		result[i].Addresses = slices.Clone(result[i].Addresses)
	}
	return result
}

// Snapshot returns a defensive copy in stable service/ID order, or an empty slice.
func (c *Cache) Snapshot() []model.Service {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return clone(c.services)
}

// Refresh replaces a snapshot only when the complete scan succeeds.
func (c *Cache) Refresh(ctx context.Context, scanner Scanner) error {
	services, err := scanner.Scan(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	unique := make(map[string]model.Service)
	for _, service := range services {
		key := service.Service + "\x00" + service.ID
		if previous, ok := unique[key]; ok {
			service.Addresses = append(slices.Clone(previous.Addresses), service.Addresses...)
		}
		service.Addresses = slices.Clone(service.Addresses)
		slices.Sort(service.Addresses)
		service.Addresses = slices.Compact(service.Addresses)
		unique[key] = service
	}
	snapshot := make([]model.Service, 0, len(unique))
	for _, service := range unique {
		snapshot = append(snapshot, service)
	}
	slices.SortFunc(snapshot, func(a, b model.Service) int {
		if order := strings.Compare(a.Service, b.Service); order != 0 {
			return order
		}
		return strings.Compare(a.ID, b.ID)
	})
	c.mu.Lock()
	c.services = snapshot
	c.mu.Unlock()
	return nil
}

// Run scans immediately and then every 30 seconds until cancellation.
func (c *Cache) Run(ctx context.Context, scanner Scanner) {
	ticker := time.NewTicker(RefreshInterval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		if err := c.Refresh(ctx, scanner); err != nil && ctx.Err() == nil {
			log.Warn().Err(err).Msg("FX discovery scan failed; keeping previous results")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
