package tracking

import (
	"math"
	"sync"

	"github.com/Kameleoon/client-go/v3/storage"
	"github.com/puzpuzpuz/xsync/v3"
)

type VisitorTrackingRegistry interface {
	Add(visitorCode string)
	AddAll(visitorCodes []string)
	// Removes and returns visitor codes; the count is bounded by `limit` and by the registry's own policy.
	Extract(limit int) []string
}

const (
	DefaultStorageLimit                           = 1_000_000
	DefaultExtractionLimit                        = 20_000
	LimitedExtractionThresholdCoefficient         = 2
	RemovalFactor                         float64 = 0.8
	UnlimitedExtraction                           = math.MaxInt
)

// ConcurrentVisitorTrackingRegistry has no registry-wide lock on the hot path: `Add` is a concurrent-map
// LoadOrStore, which is a lock-free read when the code is already registered and a single-bucket lock otherwise, so
// request goroutines rarely contend. The map is never swapped out; `Extract` removes codes from it, which guarantees
// a concurrently added code is either returned now or kept for the next extraction, never lost. Codes are extracted
// in no particular order.
type ConcurrentVisitorTrackingRegistry struct {
	visitorManager  storage.VisitorManager
	storageLimit    int
	extractionLimit int
	visitors        *xsync.MapOf[string, struct{}]

	extractMx sync.Mutex // serializes extraction and eviction; guards `snapshot`
	// Keys snapshotted from `visitors` and not yet extracted, used only for batches smaller than the extraction limit
	// (probe mode). The map has no resumable iteration and Range restarts from the first bucket, so extracting a few
	// codes per call consumes a snapshot across calls instead of rescanning the table per call.
	snapshot []string
}

func NewConcurrentVisitorTrackingRegistry(
	visitorManager storage.VisitorManager, storageLimit int, extractionLimit int,
) *ConcurrentVisitorTrackingRegistry {
	return &ConcurrentVisitorTrackingRegistry{
		visitorManager:  visitorManager,
		storageLimit:    storageLimit,
		extractionLimit: extractionLimit,
		visitors:        xsync.NewMapOf[string, struct{}](),
	}
}

func (vtr *ConcurrentVisitorTrackingRegistry) Add(visitorCode string) {
	vtr.visitors.LoadOrStore(visitorCode, struct{}{})
}

func (vtr *ConcurrentVisitorTrackingRegistry) AddAll(visitorCodes []string) {
	for _, visitorCode := range visitorCodes {
		vtr.visitors.LoadOrStore(visitorCode, struct{}{})
	}
	if vtr.size() > vtr.storageLimit {
		vtr.eraseToStorageLimit()
	}
}

func (vtr *ConcurrentVisitorTrackingRegistry) Extract(limit int) []string {
	vtr.extractMx.Lock()
	defer vtr.extractMx.Unlock()
	if limit < vtr.extractionLimit {
		return vtr.removeFromSnapshot(limit)
	}
	if limit > vtr.extractionLimit {
		size := vtr.size()
		if (size <= limit) && (size < vtr.extractionLimit*LimitedExtractionThresholdCoefficient) {
			return vtr.remove(UnlimitedExtraction)
		}
		limit = vtr.extractionLimit
	}
	return vtr.remove(limit)
}

func (vtr *ConcurrentVisitorTrackingRegistry) size() int {
	if size := vtr.visitors.Size(); size > 0 {
		return size
	}
	return 0
}

// Removes up to `count` codes in one pass over the map. Requires `extractMx` held.
func (vtr *ConcurrentVisitorTrackingRegistry) remove(count int) []string {
	vtr.snapshot = nil // codes listed by a pending snapshot are still in the map and are found by this pass
	capacity := count
	if size := vtr.size(); size < capacity {
		capacity = size
	}
	result := make([]string, 0, capacity)
	vtr.visitors.Range(func(visitorCode string, _ struct{}) bool {
		vtr.visitors.Delete(visitorCode)
		result = append(result, visitorCode)
		return len(result) < count
	})
	return result
}

// Removes up to `count` codes from the current snapshot, taking a new one only when it is used up. A batch never
// spans two snapshots, so a code cannot be returned twice in one batch; a short batch simply leaves the rest for
// the next extraction. Requires `extractMx` held.
func (vtr *ConcurrentVisitorTrackingRegistry) removeFromSnapshot(count int) []string {
	if len(vtr.snapshot) == 0 {
		vtr.snapshot = vtr.takeSnapshot()
	}
	if count > len(vtr.snapshot) {
		count = len(vtr.snapshot)
	}
	result := make([]string, 0, count)
	for (len(result) < count) && (len(vtr.snapshot) > 0) {
		visitorCode := vtr.snapshot[0]
		vtr.snapshot = vtr.snapshot[1:]
		// Already removed by eviction: the snapshot is stale
		if _, exists := vtr.visitors.LoadAndDelete(visitorCode); exists {
			result = append(result, visitorCode)
		}
	}
	return result
}

func (vtr *ConcurrentVisitorTrackingRegistry) takeSnapshot() []string {
	keys := make([]string, 0, vtr.size())
	vtr.visitors.Range(func(visitorCode string, _ struct{}) bool {
		keys = append(keys, visitorCode)
		return true
	})
	return keys
}

func (vtr *ConcurrentVisitorTrackingRegistry) eraseToStorageLimit() {
	vtr.extractMx.Lock()
	defer vtr.extractMx.Unlock()
	vtr.visitors.Range(func(visitorCode string, _ struct{}) bool {
		if vtr.visitorManager.PeekVisitor(visitorCode) == nil {
			vtr.visitors.Delete(visitorCode)
		}
		return true
	})
	visitorsToRemoveCount := vtr.size() - int(float64(vtr.storageLimit)*RemovalFactor)
	if visitorsToRemoveCount > 0 {
		vtr.remove(visitorsToRemoveCount)
	}
}
