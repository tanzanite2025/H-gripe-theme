// Package maintenancegate stops every application command before a backup or
// recovery process obtains exclusive access to application data and media.
package maintenancegate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

const FenceFilename = "application-maintenance.json"

// AcquireApplicationProcessLease must run before opening a business database,
// writing media, migrating, starting listeners or starting any scheduler.
func AcquireApplicationProcessLease(ctx context.Context, directory string, onFence func()) (*flock.Flock, error) {
	if directory == "" {
		return nil, nil
	}
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return nil, errors.New("maintenance directory must exist")
	}
	lease := flock.New(filepath.Join(directory, "application-runtime.lock"))
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		fenced, err := IsApplicationFenced(directory)
		if err != nil {
			return nil, err
		}
		if !fenced {
			locked, err := lease.TryRLock()
			if err != nil {
				return nil, err
			}
			if locked {
				fenced, err = IsApplicationFenced(directory)
				if err != nil {
					_ = lease.Close()
					return nil, err
				}
				if !fenced {
					break
				}
				if err := lease.Unlock(); err != nil {
					return nil, err
				}
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fenced, err := IsApplicationFenced(directory)
				if fenced || err != nil {
					onFence()
					return
				}
			}
		}
	}()
	return lease, nil
}

func IsApplicationFenced(directory string) (bool, error) {
	_, err := os.Stat(filepath.Join(directory, FenceFilename))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return true, fmt.Errorf("read maintenance fence: %w", err)
}

func AcquireExclusiveRecoveryLease(ctx context.Context, directory string) (*flock.Flock, error) {
	lease := flock.New(filepath.Join(directory, "application-runtime.lock"))
	locked, err := lease.TryLockContext(ctx, 200*time.Millisecond)
	if err != nil || !locked {
		_ = lease.Close()
		return nil, fmt.Errorf("application processes have not stopped: %w", err)
	}
	return lease, nil
}
