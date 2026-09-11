package storage

import (
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
	"github.com/Kameleoon/client-go/v3/managers/data"
	"github.com/Kameleoon/client-go/v3/types"
	"github.com/puzpuzpuz/xsync/v3"
)

type VisitorManager interface {
	GetVisitor(visitorCode string) Visitor
	GetOrCreateVisitor(visitorCode string) Visitor

	PeekVisitor(visitorCode string) Visitor

	AddData(visitorCode string, data ...types.Data) Visitor
	AddDataWithTrack(visitorCode string, track bool, data ...types.Data) Visitor

	Enumerate(f func(string, Visitor) bool)
	Len() int

	Clear()

	Close()
}

type VisitorManagerImpl struct {
	dataManager      data.DataManager
	visitors         *xsync.MapOf[string, *VisitorImpl]
	expirationPeriod time.Duration
	purgeTicker      *time.Ticker
	stopChan         chan struct{}
}

func NewVisitorManagerImpl(
	dataManager data.DataManager, expirationPeriod time.Duration,
) *VisitorManagerImpl {
	logging.Debug("CALL: NewVisitorManagerImpl(expirationPeriod: %s)", expirationPeriod)
	vm := &VisitorManagerImpl{
		dataManager:      dataManager,
		visitors:         xsync.NewMapOf[string, *VisitorImpl](),
		expirationPeriod: expirationPeriod,
		purgeTicker:      time.NewTicker(expirationPeriod),
		stopChan:         make(chan struct{}, 8),
	}
	go func() {
		for {
			select {
			case <-vm.purgeTicker.C:
				vm.purge()
			case <-vm.stopChan:
				return
			}
		}
	}()
	logging.Debug("RETURN: NewVisitorManagerImpl(expirationPeriod: %s) -> (VisitorManagerImpl)",
		expirationPeriod)
	return vm
}

func (vm *VisitorManagerImpl) ExpirationPeriod() time.Duration {
	return vm.expirationPeriod
}

func (vm *VisitorManagerImpl) Close() {
	vm.stop()
}
func (vm *VisitorManagerImpl) stop() {
	logging.Debug("CALL: VisitorManagerImpl.stop()")
	vm.purgeTicker.Stop()
	if len(vm.stopChan) == 0 {
		vm.stopChan <- struct{}{}
	}
	logging.Debug("RETURN: VisitorManagerImpl.stop()")
}

func (vm *VisitorManagerImpl) GetVisitor(visitorCode string) Visitor {
	logging.Debug("CALL: VisitorManagerImpl.GetVisitor(visitorCode: %s)", visitorCode)
	var visitor Visitor
	// Compute runs the callback under the map's bucket lock, so the activity time is updated before `purge`
	// can decide to remove the visitor. A missing visitor is left missing (delete of an absent key is a no-op).
	if v, ok := vm.visitors.Compute(visitorCode, func(v *VisitorImpl, loaded bool) (*VisitorImpl, bool) {
		if loaded {
			v.UpdateLastActivityTime()
		}
		return v, !loaded
	}); ok {
		visitor = v
	}
	logging.Debug("RETURN: VisitorManagerImpl.GetVisitor(visitorCode: %s) -> (visitor: %s)",
		visitorCode, visitor)
	return visitor
}

func (vm *VisitorManagerImpl) GetOrCreateVisitor(visitorCode string) Visitor {
	return vm.getOrCreateVisitor(visitorCode)
}
func (vm *VisitorManagerImpl) getOrCreateVisitor(visitorCode string) *VisitorImpl {
	logging.Debug("CALL: VisitorManagerImpl.getOrCreateVisitor(visitorCode: %s)", visitorCode)
	visitor, _ := vm.visitors.Compute(visitorCode, func(former *VisitorImpl, loaded bool) (*VisitorImpl, bool) {
		if loaded {
			former.UpdateLastActivityTime()
			return former, false
		}
		return NewVisitorImpl(), false
	})
	logging.Debug("RETURN: VisitorManagerImpl.getOrCreateVisitor(visitorCode: %s) -> (visitor)", visitorCode)
	return visitor
}

func (vm *VisitorManagerImpl) PeekVisitor(visitorCode string) Visitor {
	logging.Debug("CALL: VisitorManagerImpl.PeekVisitor(visitorCode: %s)", visitorCode)

	var visitor Visitor
	if v, ok := vm.visitors.Load(visitorCode); ok {
		visitor = v
	}

	logging.Debug("RETURN: VisitorManagerImpl.PeekVisitor(visitorCode: %s) -> (visitor: %s)",
		visitorCode, visitor)

	return visitor
}

func (vm *VisitorManagerImpl) AddData(visitorCode string, data ...types.Data) Visitor {
	return vm.AddDataWithTrack(visitorCode, true, data...)
}

func (vm *VisitorManagerImpl) AddDataWithTrack(visitorCode string, track bool, data ...types.Data) Visitor {
	logging.Debug("CALL: VisitorManagerImpl.AddDataWithTrack(visitorCode: %s, track: %t, data: %s)", visitorCode, track, data)
	visitor := vm.getOrCreateVisitor(visitorCode)
	cdi := vm.dataManager.DataFile().CustomDataInfo()
	if cdi != nil {
		for i, d := range data {
			switch dataImpl := d.(type) {
			case *types.CustomData:
				data[i] = vm.processCustomData(visitorCode, visitor, cdi, dataImpl)
			case *types.Conversion:
				data[i] = vm.processConversion(dataImpl, cdi)
			}
			if !track {
				if sendable, ok := data[i].(types.Sendable); ok {
					sendable.MarkAsSent()
				}
			}
		}
	}
	visitor.AddData(data...)
	logging.Debug("RETURN: VisitorManagerImpl.AddDataWithTrack(visitorCode: %s, track: %t, data: %s) -> (visitor)", visitorCode, track, data)
	return visitor
}

func (vm *VisitorManagerImpl) processCustomData(
	visitorCode string,
	visitor *VisitorImpl,
	cdi *types.CustomDataInfo,
	cd *types.CustomData,
) types.Data {
	if mappedCd := vm.tryMapCustomDataIndexByName(cd, cdi); mappedCd == nil {
		logging.Error("%s is invalid and will be ignored", cd)
		return nil
	} else {
		cd = mappedCd
	}
	// We shouldn't send custom data with local only type
	if cdi.IsLocalOnly(cd.Index()) {
		cd.MarkAsSent()
	}
	// If mappingIdentifier is passed, we should link anonymous visitor with real unique userId.
	// After authorization, customer must be able to continue work with userId, but hash for variation
	// should be calculated based on anonymous visitor code, that's why set MappingIdentifier to visitor.
	if isMappingIdentifier(cdi, cd) {
		visitor.SetMappingIdentifier(&visitorCode)
		userId := cd.Values()[0]
		if visitorCode != userId {
			vm.visitors.Store(userId, cloneVisitorImpl(visitor))
			logging.Info("Linked anonymous visitor '%s' with user '%s'", visitorCode, userId)
		}
		return types.NewMappingIdentifier(cd)
	}
	return cd
}

func (vm *VisitorManagerImpl) processConversion(c *types.Conversion, cdi *types.CustomDataInfo) types.Data {
	metadata := c.Metadata()
	for i, cd := range metadata {
		cd = vm.tryMapCustomDataIndexByName(cd, cdi)
		if cd == nil {
			logging.Warning("Conversion metadata %s is invalid", metadata[i])
		}
		metadata[i] = cd
	}
	return c
}

func (vm *VisitorManagerImpl) tryMapCustomDataIndexByName(
	cd *types.CustomData,
	cdi *types.CustomDataInfo,
) *types.CustomData {
	if cd.Index() != types.CustomDataUndefinedIndex {
		return cd
	}
	if cd.Name() == "" {
		return nil
	}
	cdIndex, exists := cdi.GetCustomDataIndexByName(cd.Name())
	if exists {
		return cd.NamedToIndexed(cdIndex)
	}
	return nil
}

func isMappingIdentifier(cdi *types.CustomDataInfo, cd types.ICustomData) bool {
	return (cd.Index() == cdi.MappingIdentifierIndex()) && (len(cd.Values()) > 0) && (cd.Values()[0] != "")
}

func (vm *VisitorManagerImpl) Enumerate(f func(string, Visitor) bool) {
	vm.visitors.Range(func(visitorCode string, visitor *VisitorImpl) bool {
		return f(visitorCode, visitor)
	})
}
func (vm *VisitorManagerImpl) Len() int {
	return vm.visitors.Size()
}

func (vm *VisitorManagerImpl) purge() {
	logging.Debug("CALL: VisitorManagerImpl.purge()")
	expiredDT := time.Now().Add(-vm.expirationPeriod)
	vm.visitors.Range(func(visitorCode string, v *VisitorImpl) bool {
		if v.LastActivityTime().Before(expiredDT) {
			// re-checked under the bucket lock: the visitor may have been touched since the Range read it
			vm.visitors.Compute(visitorCode, func(v *VisitorImpl, loaded bool) (*VisitorImpl, bool) {
				return v, loaded && v.LastActivityTime().Before(expiredDT)
			})
		}
		return true
	})
	logging.Debug("RETURN: VisitorManagerImpl.purge()")
}

func (vm *VisitorManagerImpl) Clear() {
	logging.Debug("CALL: VisitorManagerImpl.Clear()")
	vm.visitors.Clear()
	logging.Debug("RETURN: VisitorManagerImpl.Clear()")
}
