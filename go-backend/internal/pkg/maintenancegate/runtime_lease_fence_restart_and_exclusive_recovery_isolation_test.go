package maintenancegate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecoveryWaitsForEveryRuntimeLeaseAndBlocksRestartBeforeDatabaseAccess(t *testing.T) {
	directory := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan struct{}, 1)
	lease, err := AcquireApplicationProcessLease(ctx, directory, func() { stopped <- struct{}{} })
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, FenceFilename), []byte(`{"task_id":"isolated"}`), 0600))
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("runtime did not observe maintenance fence")
	}
	deadline, finish := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer finish()
	_, err = AcquireExclusiveRecoveryLease(deadline, directory)
	require.Error(t, err)
	require.NoError(t, lease.Close())
	exclusive, err := AcquireExclusiveRecoveryLease(context.Background(), directory)
	require.NoError(t, err)
	defer exclusive.Close()
	deadline2, finish2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer finish2()
	_, err = AcquireApplicationProcessLease(deadline2, directory, func() {})
	require.Error(t, err)
}
