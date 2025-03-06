package backup

import (
	"testing"
)

func TestNewBackupStats(t *testing.T) {
	stats := NewBackupStats()

	if stats.ResourceCount != 0 {
		t.Errorf("Expected ResourceCount to be 0, got %d", stats.ResourceCount)
	}
	if stats.ErrorCount != 0 {
		t.Errorf("Expected ErrorCount to be 0, got %d", stats.ErrorCount)
	}
	if len(stats.ResourcesBackedUp) != 0 {
		t.Errorf("Expected ResourcesBackedUp to be empty, got %v", stats.ResourcesBackedUp)
	}
	if len(stats.ResourceErrors) != 0 {
		t.Errorf("Expected ResourceErrors to be empty, got %v", stats.ResourceErrors)
	}
}

// Test that the BackupStats.ResourcesBackedUp map is updated correctly
func TestResourcesBackedUpTracking(t *testing.T) {
	stats := NewBackupStats()

	// Manually update stats as if resources were backed up
	stats.ResourceCount += 5
	stats.ResourcesBackedUp["Pod"] = 2
	stats.ResourcesBackedUp["Service"] = 3

	// Verify counts
	if stats.ResourceCount != 5 {
		t.Errorf("Expected ResourceCount to be 5, got %d", stats.ResourceCount)
	}

	if stats.ResourcesBackedUp["Pod"] != 2 {
		t.Errorf("Expected ResourcesBackedUp[Pod] to be 2, got %d", stats.ResourcesBackedUp["Pod"])
	}

	if stats.ResourcesBackedUp["Service"] != 3 {
		t.Errorf("Expected ResourcesBackedUp[Service] to be 3, got %d", stats.ResourcesBackedUp["Service"])
	}
}
