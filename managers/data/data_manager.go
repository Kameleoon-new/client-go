package data

import (
	"github.com/Kameleoon/client-go/v3/types"
)

type DataManager interface {
	DataFile() types.IDataFile
	ExternalDataFile() *types.DataFile
	IsVisitorCodeManaged() bool

	SetDataFile(dataFile types.IDataFile)
}

type DataManagerImpl struct {
	container *dataContainer
}

func NewDataManagerImpl(dataFile types.IDataFile) *DataManagerImpl {
	dm := &DataManagerImpl{}
	dm.SetDataFile(dataFile)
	return dm
}

func (dm *DataManagerImpl) ExternalDataFile() *types.DataFile {
	return dm.container.externalDataFile
}

func (dm *DataManagerImpl) DataFile() types.IDataFile {
	return dm.container.dataFile
}

func (dm *DataManagerImpl) IsVisitorCodeManaged() bool {
	return dm.container.isVisitorCodeManaged
}

func (dm *DataManagerImpl) SetDataFile(dataFile types.IDataFile) {
	dm.container = newDataContainer(dataFile)
}

type dataContainer struct {
	dataFile             types.IDataFile
	externalDataFile     *types.DataFile
	isVisitorCodeManaged bool
}

func newDataContainer(dataFile types.IDataFile) *dataContainer {
	return &dataContainer{
		dataFile:             dataFile,
		externalDataFile:     (types.DataFile{}).BuildFromInternal(dataFile),
		isVisitorCodeManaged: dataFile.Settings().IsConsentRequired() && !dataFile.HasAnyTargetedDeliveryRule(),
	}
}
