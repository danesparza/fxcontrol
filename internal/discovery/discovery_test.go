package discovery

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/danesparza/fxcontrol/internal/model"
	"github.com/stretchr/testify/mock"
)

func TestRefresh(t *testing.T) {
	old := model.Service{Service: "fxaudio", ID: "old", Addresses: []string{"192.168.1.1"}}
	next := model.Service{Service: "fxpixel", ID: "new", Addresses: []string{"192.168.1.2"}}
	for _, tc := range []struct {
		name    string
		results []model.Service
		err     error
		want    []model.Service
	}{
		{"replace removes missing", []model.Service{next}, nil, []model.Service{next}},
		{"empty clears stale", nil, nil, []model.Service{}},
		{"failure retains previous", nil, errors.New("network unavailable"), []model.Service{old}},
		{"deduplicate and sort", []model.Service{next, old, next}, nil, []model.Service{old, next}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := &Cache{services: []model.Service{old}}
			scanner := NewMockScanner(t)
			scanner.EXPECT().Scan(t.Context()).Return(tc.results, tc.err).Once()
			err := cache.Refresh(t.Context(), scanner)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got error %v, want %v", err, tc.err)
			}
			if got := cache.Snapshot(); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestSnapshotOwnership(t *testing.T) {
	cache := &Cache{}
	if got := cache.Snapshot(); got == nil || len(got) != 0 {
		t.Fatalf("empty snapshot: %v", got)
	}
	input := []model.Service{{ID: "one", Addresses: []string{"192.168.1.1"}}, {ID: "one", Addresses: []string{"192.168.1.2"}}}
	scanner := NewMockScanner(t)
	scanner.EXPECT().Scan(t.Context()).Return(input, nil).Once()
	if err := cache.Refresh(t.Context(), scanner); err != nil {
		t.Fatal(err)
	}
	input[0].Addresses[0] = "modified"
	snapshot := cache.Snapshot()
	if !reflect.DeepEqual(snapshot[0].Addresses, []string{"192.168.1.1", "192.168.1.2"}) {
		t.Fatalf("addresses not merged/copied: %v", snapshot)
	}
	snapshot[0].ID = "modified"
	snapshot[0].Addresses[0] = "modified"
	if got := cache.Snapshot(); got[0].ID != "one" || got[0].Addresses[0] != "192.168.1.1" {
		t.Fatalf("snapshot shares cache memory: %v", got)
	}
}

func TestRunAndConcurrentSnapshot(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	cache := &Cache{services: []model.Service{{ID: "previous"}}}
	scanner := NewMockScanner(t)
	started := make(chan struct{})
	scanner.EXPECT().Scan(mock.Anything).RunAndReturn(func(ctx context.Context) ([]model.Service, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	}).Once()
	stopped := make(chan struct{})
	go func() { defer close(stopped); cache.Run(ctx, scanner) }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("initial scan did not start")
	}
	read := make(chan []model.Service, 1)
	go func() { read <- cache.Snapshot() }()
	select {
	case got := <-read:
		if len(got) != 1 || got[0].ID != "previous" {
			t.Fatalf("unexpected snapshot: %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("snapshot blocked on active scan")
	}
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("discovery did not stop")
	}
}

func TestConcurrentRefreshAndRead(t *testing.T) {
	cache := &Cache{}
	scanner := NewMockScanner(t)
	scanner.EXPECT().Scan(t.Context()).Return([]model.Service{{ID: "one", Addresses: []string{"192.168.1.1"}}}, nil).Times(100)
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			if err := cache.Refresh(t.Context(), scanner); err != nil {
				t.Error(err)
			}
		}
	})
	wg.Go(func() {
		for range 100 {
			snapshot := cache.Snapshot()
			if len(snapshot) > 0 {
				snapshot[0].Addresses[0] = "changed"
			}
		}
	})
	wg.Wait()
}

func TestCanceledRefreshPreservesSnapshot(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cache := &Cache{services: []model.Service{{ID: "previous"}}}
	scanner := NewMockScanner(t)
	scanner.EXPECT().Scan(ctx).Return([]model.Service{}, nil).Once()
	if err := cache.Refresh(ctx, scanner); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if got := cache.Snapshot(); len(got) != 1 || got[0].ID != "previous" {
		t.Fatalf("cancellation replaced snapshot: %v", got)
	}
	if _, err := (NetworkScanner{}).Scan(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestRunRefreshesEveryThirtySeconds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		cache := &Cache{}
		scanner := NewMockScanner(t)
		scanner.EXPECT().Scan(ctx).Return([]model.Service{{ID: "first"}}, nil).Once()
		go cache.Run(ctx, scanner)
		synctest.Wait()
		if got := cache.Snapshot(); len(got) != 1 || got[0].ID != "first" {
			t.Fatalf("initial snapshot: %v", got)
		}
		scanner.EXPECT().Scan(ctx).Return([]model.Service{{ID: "second"}}, nil).Once()
		time.Sleep(30 * time.Second)
		synctest.Wait()
		if got := cache.Snapshot(); len(got) != 1 || got[0].ID != "second" {
			t.Fatalf("second snapshot: %v", got)
		}
		scanner.EXPECT().Scan(ctx).Return([]model.Service{}, nil).Once()
		time.Sleep(30 * time.Second)
		synctest.Wait()
		if got := cache.Snapshot(); len(got) != 0 {
			t.Fatalf("stale snapshot: %v", got)
		}
		cancel()
		synctest.Wait()
	})
}
