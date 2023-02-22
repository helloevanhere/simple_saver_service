package estimate

import (
	"fmt"
)

type EstimatedSavings struct {
	CalculatedMonthlylSavingsMin float64 `json:"estimated_savings_min"`
	CalculatedMonthlySavingsMax  float64 `json:"estimated_savings_max"`
}

// The pricing of storage classes per GB per month
var StorageClassPrices = map[string]float64{
	"STANDARD":     0.023,
	"STANDARD_IA":  0.0125,
	"ONEZONE_IA":   0.01,
	"GLACIER":      0.004,
	"DEEP_ARCHIVE": 0.00099,
}

// Calculate current monthly storage cost
func CurrentStorageCost(bucketSizeinBytes int64, storageClass string) (float64, error) {
	price, ok := StorageClassPrices[storageClass]
	if !ok {
		return 0, fmt.Errorf("invalid storage class: %s", storageClass)
	}

	return float64(bucketSizeinBytes) / 1024 / 1024 / 1024 * price, nil
}

func SavingsForBytesDeletedByStorageClass(objectSizeInBytes int64, storageClass string) (savings float64) {
	// Calculate the storage cost of the objects to be deleted
	monthlyCost := float64(objectSizeInBytes) / (1024 * 1024 * 1024) * StorageClassPrices[storageClass]

	return monthlyCost
}

// compressionType can be one of "gzip", "zlib", "zip" ,"snappy"
func estimateCompressedSize(dataSize int64, compressionType string) (minSize int64, maxSize int64, err error) {
	switch compressionType {
	case ".gzip":
		// GZIP compression ratio of around 2.5:1 to 4:1
		minSize = dataSize / 4
		maxSize = dataSize / 2
	case ".zlib", ".zip":
		// ZLIB, ZIP compression ratio of around 2:1 to 3:1
		minSize = dataSize / 2
		maxSize = (dataSize * 2) / 3
	case ".snappy":
		// SNAPPY compression ratio 2.5:1 to 4:1
		minSize = dataSize / 4
		maxSize = dataSize / 2
	case ".jpeg":
		// JPEG compression ratio of around 10:1 to 20:1
		minSize = dataSize / 20
		maxSize = dataSize / 10
	case ".mp3":
		// MP3 compression ratio of around 10:1 to 12:1
		minSize = dataSize / 12
		maxSize = dataSize / 10
	case ".h264":
		// H.264 compression ratio of around 10:1 to 20:1
		minSize = dataSize / 20
		maxSize = dataSize / 10
	case ".bzip2":
		// BZIP2 compression ratio of around 1.5:1 to 3:1
		minSize = (dataSize * 2) / 3
		maxSize = int64(float64(dataSize) / 1.5)
	case ".zstd":
		// ZSTD compression ratio of around 2.5:1 to 4:1
		minSize = dataSize / 4
		maxSize = dataSize / 2
	case ".7zip":
		// 7ZIP compression ratio of around 2:1 to 3:1
		minSize = dataSize / 3
		maxSize = dataSize / 2
	default:
		err = fmt.Errorf("error estimating compressed size, unsupported compression type: %s", compressionType)
	}
	return minSize, maxSize, err
}

func SavingsForBytesCompressedByStorageClass(dataSize int64, compressionType string, storageClass string) (float64, float64, error) {
	minSize, maxSize, err := estimateCompressedSize(dataSize, compressionType)
	if err != nil {
		return 0, 0, err
	}

	price, ok := StorageClassPrices[storageClass]
	if !ok {
		return 0, 0, fmt.Errorf("invalid storage class: %s", storageClass)
	}

	minSaving := float64(maxSize) / 1024 / 1024 / 1024 * price
	maxSaving := float64(minSize) / 1024 / 1024 / 1024 * price

	return minSaving, maxSaving, nil
}
