package ordernumber

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGeneratorProducesOpaqueValidatedNumber(t *testing.T) {
	generator, err := NewGenerator("test-order-number-secret", 3)
	if err != nil {
		t.Fatal(err)
	}

	value, err := generator.GenerateAt(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(value, "TZ-2026-") {
		t.Fatalf("order number = %q, want 2026 prefix", value)
	}
	if !IsProtectedFormat(value) || !generator.Validate(value) {
		t.Fatalf("generated order number should validate: %q", value)
	}
	if strings.Contains(value, "000000") {
		t.Fatalf("order number appears to expose an internal sequence: %q", value)
	}
}

func TestGeneratorCreatesUniqueNumbersWithinSameMillisecond(t *testing.T) {
	generator, err := NewGenerator("test-order-number-secret", 1)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		value, err := generator.GenerateAt(now)
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := seen[value]; exists {
			t.Fatalf("duplicate order number generated: %s", value)
		}
		seen[value] = struct{}{}
	}
}

func TestGeneratorCreatesUniqueNumbersConcurrently(t *testing.T) {
	generator, err := NewGenerator("test-order-number-secret", 1)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 16
	const numbersPerWorker = 128
	seen := make(map[string]struct{}, workers*numbersPerWorker)
	var seenMu sync.Mutex
	var waitGroup sync.WaitGroup
	errs := make(chan error, workers)

	for worker := 0; worker < workers; worker++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for index := 0; index < numbersPerWorker; index++ {
				value, generateErr := generator.Generate()
				if generateErr != nil {
					errs <- generateErr
					return
				}
				seenMu.Lock()
				_, exists := seen[value]
				seen[value] = struct{}{}
				seenMu.Unlock()
				if exists {
					errs <- ErrInvalidOrderNo
					return
				}
			}
		}()
	}

	waitGroup.Wait()
	close(errs)
	for generateErr := range errs {
		t.Fatalf("concurrent generation failed: %v", generateErr)
	}
	if len(seen) != workers*numbersPerWorker {
		t.Fatalf("generated %d unique order numbers, want %d", len(seen), workers*numbersPerWorker)
	}
}

func TestGeneratorRejectsTamperedChecksum(t *testing.T) {
	generator, err := NewGenerator("test-order-number-secret", 0)
	if err != nil {
		t.Fatal(err)
	}

	value, err := generator.GenerateAt(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	tampered := value[:len(value)-1] + "A"
	if tampered == value {
		tampered = value[:len(value)-1] + "B"
	}
	if generator.Validate(tampered) {
		t.Fatalf("tampered order number should be rejected: %s", tampered)
	}
}

func TestIsSequentialCandidateDetectsAutoIncrementNumbers(t *testing.T) {
	for _, value := range []string{"1001", "#1001", " 1001 "} {
		if !IsSequentialCandidate(value) {
			t.Fatalf("%q should be treated as a sequential order number candidate", value)
		}
	}
	for _, value := range []string{"ORD-1001", "TZ-2026-ABCDEFGHIJKLMNOPQRST"} {
		if IsSequentialCandidate(value) {
			t.Fatalf("%q should not be treated as a sequential order number candidate", value)
		}
	}
}

func TestGeneratorAcceptsPreviousSecretDuringRotation(t *testing.T) {
	legacyGenerator, err := NewGenerator("legacy-order-number-secret", 2)
	if err != nil {
		t.Fatal(err)
	}
	legacyNumber, err := legacyGenerator.GenerateAt(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	rotatedGenerator, err := NewGeneratorWithPreviousSecret("current-order-number-secret", "legacy-order-number-secret", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !rotatedGenerator.Validate(legacyNumber) {
		t.Fatalf("previous-secret order number should validate after rotation: %s", legacyNumber)
	}

	currentNumber, err := rotatedGenerator.GenerateAt(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !rotatedGenerator.Validate(currentNumber) {
		t.Fatalf("current-secret order number should validate after rotation: %s", currentNumber)
	}
	if legacyGenerator.Validate(currentNumber) {
		t.Fatalf("old secret must not validate new order number: %s", currentNumber)
	}
}

func TestGeneratorRejectsInvalidConfiguration(t *testing.T) {
	if _, err := NewGenerator("", 0); err != ErrMissingSecret {
		t.Fatalf("missing secret error = %v, want %v", err, ErrMissingSecret)
	}
	if _, err := NewGenerator("secret", maxNodeID+1); err != ErrInvalidNodeID {
		t.Fatalf("node id error = %v, want %v", err, ErrInvalidNodeID)
	}
}
