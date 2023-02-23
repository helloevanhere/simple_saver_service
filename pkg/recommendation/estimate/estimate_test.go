package estimate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentStorageCost(t *testing.T) {
	tests := []struct {
		bucketSizeInBytes int64
		storageClass      string
		expectedCost      float64
	}{
		{0, "STANDARD", 0},
		{1000000000, "STANDARD", 0.023},
		{1000000000, "STANDARD_IA", 0.0125},
		{1000000000, "ONEZONE_IA", 0.01},
		{1000000000, "GLACIER", 0.004},
		{1000000000, "DEEP_ARCHIVE", 0.00099},
		{2000000000, "STANDARD", 0.046},
	}

	for _, test := range tests {
		cost, err := CurrentStorageCost(test.bucketSizeInBytes, test.storageClass)
		assert.NoError(t, err)
		assert.InDelta(t, test.expectedCost, cost, 0.001)
	}
}

func TestSavingsForBytesDeletedByStorageClass(t *testing.T) {
	tests := []struct {
		objectSizeInBytes int64
		storageClass      string
		expectedSavings   float64
	}{
		{0, "STANDARD", 0},
		{1000000000, "STANDARD", 0.023},
		{1000000000, "STANDARD_IA", 0.0125},
		{1000000000, "ONEZONE_IA", 0.01},
		{1000000000, "GLACIER", 0.004},
		{1000000000, "DEEP_ARCHIVE", 0.00099},
		{2000000000, "STANDARD", 0.046},
	}

	for _, test := range tests {
		savings := SavingsForBytesDeletedByStorageClass(test.objectSizeInBytes, test.storageClass)
		assert.InDelta(t, test.expectedSavings, savings, 0.001)
	}
}

func TestSavingsForBytesCompressedByStorageClass(t *testing.T) {
	tests := []struct {
		dataSize        int64
		compressionType string
		storageClass    string
		expectedMin     float64
		expectedMax     float64
	}{
		{0, ".gzip", "STANDARD", 0, 0},
		{1000000000, ".gzip", "STANDARD", 0.0092, 0.092},
		{1000000000, ".zlib", "STANDARD", 0.015, 0.03},
		{1000000000, ".snappy", "STANDARD", 0.0092, 0.037},
		{1000000000, ".bzip2", "STANDARD", 0.015, 0.03},
		{1000000000, ".zstd", "STANDARD", 0.0092, 0.092},
		{1000000000, ".7zip", "STANDARD", 0.0115, 0.046},
		{2000000000, ".gzip", "STANDARD", 0.0184, 0.184},
	}

	for _, test := range tests {
		min, max, err := SavingsForBytesCompressedByStorageClass(test.dataSize, test.compressionType, test.storageClass)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, min, test.expectedMin)
		assert.LessOrEqual(t, max, test.expectedMax)
	}
}
