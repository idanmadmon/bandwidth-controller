package bandwidthController

import (
	"io"
	"sync"

	"github.com/google/uuid"
)

type BandwidthController struct {
	files     map[uuid.UUID]*File
	mu        sync.Mutex
	bandwidth int64
}

func NewBandwidthController(bandwidth int64) *BandwidthController {
	return &BandwidthController{
		files:     make(map[uuid.UUID]*File),
		bandwidth: bandwidth,
	}
}

func (bc *BandwidthController) GetFileReader(r io.Reader, fileSize int64) *FileReader {
	fileID := uuid.New()
	fileReader := NewFileReader(r, bc.bandwidth, func() {
		bc.removeFile(fileID)
	})

	file := NewFile(fileReader, fileSize)

	bc.mu.Lock()
	bc.files[fileID] = file
	bc.updateLimits()
	bc.mu.Unlock()

	return fileReader
}

func (bc *BandwidthController) removeFile(fileID uuid.UUID) {
	bc.mu.Lock()
	delete(bc.files, fileID)
	bc.updateLimits()
	bc.mu.Unlock()
}

func (bc *BandwidthController) updateLimits() {
	totalWeight := 0.0
	weights := make(map[uuid.UUID]float64)

	for id, file := range bc.files {
		remainingSize := file.Size - file.Reader.BytesRead()
		if remainingSize > 0 {
			weights[id] = 1.0 / float64(remainingSize)
			totalWeight += weights[id]
		}
	}

	for id, weight := range weights {
		ratio := weight / totalWeight
		newLimit := int64(float64(bc.bandwidth) * ratio)
		bc.files[id].Reader.UpdateLimit(newLimit)
	}
}
