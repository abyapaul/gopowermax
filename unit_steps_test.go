/*
 Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
package pmax

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/dell/gopowermax/mock"
	types "github.com/dell/gopowermax/types/v90"
)

const (
	defaultUsername        = "username"
	defaultPassword        = "password"
	symID                  = "000197900046"
	testPortGroup          = "12se0042-iscsi-PG"
	testInitiator          = "SE-1E:000:iqn.1993-08.org.debian:01:5ae293b352a2"
	testInitiatorIQN       = "iqn.1993-08.org.debian:01:5ae293b352a2"
	testUpdateInitiatorIQN = "iqn.1993-08.org.debian:01:5ae293b352a3"
	testUpdateInitiator    = "SE-1E:000:iqn.1993-08.org.debian:01:5ae293b352a3"
	testHost               = "l2se0042_iscsi_ig"
	testHostGroup          = "l2se0042_43_iscsi_ig"
	testSG                 = "l2se0042_sg"
	mvID                   = "12se0042_mv"
	testFCInitiatorWWN     = "10000090fa66060a"
	testFCInitiator        = "FA-1D:4:10000090fa66060a"
)

type uMV struct {
	maskingViewID  string
	hostID         string
	hostGroupID    string
	storageGroupID string
	portGroupID    string
}

type unitContext struct {
	nGoRoutines int
	client      Pmax
	err         error // First error observed

	symIDList          *types.SymmetrixIDList
	sym                *types.Symmetrix
	vol                *types.Volume
	volList            []string
	storageGroup       *types.StorageGroup
	storageGroupIDList *types.StorageGroupIDList
	jobIDList          []string
	job                *types.Job
	storagePoolList    *types.StoragePoolList
	portGroupList      *types.PortGroupList
	portGroup          *types.PortGroup
	initiatorList      *types.InitiatorList
	initiator          *types.Initiator
	hostList           *types.HostList
	host               *types.Host
	maskingViewList    *types.MaskingViewList
	maskingView        *types.MaskingView
	uMaskingView       *uMV
	addressList        []string
	storagePool        *types.StoragePool
	volIDList          []string
	hostID             string
	hostGroupID        string
	sgID               string

	symRepCapibilities    *types.SymReplicationCapabilities
	sourceVolumeList      []types.VolumeList
	symVolumeList         *types.SymVolumeList
	volSnapList           *types.SnapshotVolumeGeneration
	volumeSnapshot        *types.VolumeSnapshot
	volSnapGenerationList *types.VolumeSnapshotGenerations
	volSnapGenerationInfo *types.VolumeSnapshotGeneration
	volResultPrivate      *types.VolumeResultPrivate

	inducedErrors struct {
		badCredentials bool
		badPort        bool
		badIP          bool
	}
}

func (c *unitContext) reset() {
	Debug = true
	c.err = nil
	c.symIDList = nil
	c.sym = nil
	c.vol = nil
	c.volList = make([]string, 0)
	c.storageGroup = nil
	c.storageGroupIDList = nil
	c.portGroupList = nil
	c.portGroup = nil
	c.initiatorList = nil
	c.initiator = nil
	c.hostList = nil
	c.host = nil
	c.jobIDList = nil
	c.job = nil
	c.storagePoolList = nil
	c.maskingViewList = nil
	c.uMaskingView = nil
	c.maskingView = nil
	c.storagePool = nil
	MAXJobRetryCount = 5
	c.volIDList = make([]string, 0)
	c.hostID = ""
	c.hostGroupID = ""
	c.sgID = ""

	c.symRepCapibilities = nil
	c.sourceVolumeList = make([]types.VolumeList, 0)
	c.symVolumeList = nil
	c.volSnapList = nil
	c.volumeSnapshot = nil
	c.volSnapGenerationList = nil
	c.volSnapGenerationInfo = nil
	c.volResultPrivate = nil

}

func (c *unitContext) iInduceError(errorType string) error {
	mock.InducedErrors.InvalidJSON = false
	mock.InducedErrors.BadHTTPStatus = 0
	mock.InducedErrors.GetSymmetrixError = false
	mock.InducedErrors.GetVolumeIteratorError = false
	mock.InducedErrors.GetVolumeError = false
	mock.InducedErrors.UpdateVolumeError = false
	mock.InducedErrors.DeleteVolumeError = false
	mock.InducedErrors.DeviceInSGError = false
	mock.InducedErrors.GetStorageGroupError = false
	mock.InducedErrors.InvalidResponse = false
	mock.InducedErrors.UpdateStorageGroupError = false
	mock.InducedErrors.GetJobError = false
	mock.InducedErrors.JobFailedError = false
	mock.InducedErrors.VolumeNotCreatedError = false
	mock.InducedErrors.GetJobCannotFindRoleForUser = false
	mock.InducedErrors.CreateStorageGroupError = false
	mock.InducedErrors.StorageGroupAlreadyExists = false
	mock.InducedErrors.DeleteStorageGroupError = false
	mock.InducedErrors.GetStoragePoolListError = false
	mock.InducedErrors.GetMaskingViewError = false
	mock.InducedErrors.GetPortGroupError = false
	mock.InducedErrors.GetInitiatorError = false
	mock.InducedErrors.GetHostError = false
	mock.InducedErrors.MaskingViewAlreadyExists = false
	mock.InducedErrors.DeleteMaskingViewError = false
	mock.InducedErrors.CreateMaskingViewError = false
	mock.InducedErrors.PortGroupNotFoundError = false
	mock.InducedErrors.InitiatorGroupNotFoundError = false
	mock.InducedErrors.StorageGroupNotFoundError = false
	mock.InducedErrors.CreateHostError = false
	mock.InducedErrors.DeleteHostError = false
	mock.InducedErrors.VolumeNotAddedError = false
	mock.InducedErrors.UpdateHostError = false
	mock.InducedErrors.GetPortError = false
	mock.InducedErrors.GetDirectorError = false
	mock.InducedErrors.GetStoragePoolError = false

	switch errorType {
	case "InvalidJSON":
		mock.InducedErrors.InvalidJSON = true
	case "httpStatus500":
		mock.InducedErrors.BadHTTPStatus = 500
	case "GetSymmetrixError":
		mock.InducedErrors.GetSymmetrixError = true
	case "GetVolumeIteratorError":
		mock.InducedErrors.GetVolumeIteratorError = true
	case "GetVolumeError":
		mock.InducedErrors.GetVolumeError = true
	case "UpdateVolumeError":
		mock.InducedErrors.UpdateVolumeError = true
	case "DeleteVolumeError":
		mock.InducedErrors.DeleteVolumeError = true
	case "DeviceInSGError":
		mock.InducedErrors.DeviceInSGError = true
	case "GetStorageGroupError":
		mock.InducedErrors.GetStorageGroupError = true
	case "InvalidResponse":
		mock.InducedErrors.InvalidResponse = true
	case "UpdateStorageGroupError":
		mock.InducedErrors.UpdateStorageGroupError = true
	case "GetJobError":
		mock.InducedErrors.GetJobError = true
	case "JobFailedError":
		mock.InducedErrors.JobFailedError = true
	case "VolumeNotCreatedError":
		mock.InducedErrors.VolumeNotCreatedError = true
	case "GetJobCannotFindRoleForUser":
		mock.InducedErrors.GetJobCannotFindRoleForUser = true
	case "CreateStorageGroupError":
		mock.InducedErrors.CreateStorageGroupError = true
	case "StorageGroupAlreadyExists":
		mock.InducedErrors.StorageGroupAlreadyExists = true
	case "DeleteStorageGroupError":
		mock.InducedErrors.DeleteStorageGroupError = true
	case "GetStoragePoolListError":
		mock.InducedErrors.GetStoragePoolListError = true
	case "GetMaskingViewError":
		mock.InducedErrors.GetMaskingViewError = true
	case "GetPortGroupError":
		mock.InducedErrors.GetPortGroupError = true
	case "GetInitiatorError":
		mock.InducedErrors.GetInitiatorError = true
	case "GetHostError":
		mock.InducedErrors.GetHostError = true
	case "CreateMaskingViewError":
		mock.InducedErrors.CreateMaskingViewError = true
	case "MaskingViewAlreadyExists":
		mock.InducedErrors.MaskingViewAlreadyExists = true
	case "DeleteMaskingViewError":
		mock.InducedErrors.DeleteMaskingViewError = true
	case "PortGroupNotFoundError":
		mock.InducedErrors.PortGroupNotFoundError = true
	case "InitiatorGroupNotFoundError":
		mock.InducedErrors.InitiatorGroupNotFoundError = true
	case "StorageGroupNotFoundError":
		mock.InducedErrors.StorageGroupNotFoundError = true
	case "CreateHostError":
		mock.InducedErrors.CreateHostError = true
	case "DeleteHostError":
		mock.InducedErrors.DeleteHostError = true
	case "VolumeNotAddedError":
		mock.InducedErrors.VolumeNotAddedError = true
	case "UpdateHostError":
		mock.InducedErrors.UpdateHostError = true
	case "GetPortError":
		mock.InducedErrors.GetPortError = true
	case "GetDirectorError":
		mock.InducedErrors.GetDirectorError = true
	case "GetStoragePoolError":
		mock.InducedErrors.GetStoragePoolError = true
	case "GetSymVolumeError":
		mock.InducedErrors.GetSymVolumeError = true
	case "GetSnapshotError":
		mock.InducedErrors.GetSnapshotError = true
	case "GetGenerationError":
		mock.InducedErrors.GetGenerationError = true
	case "GetPrivateVolumeIterator":
		mock.InducedErrors.GetPrivateVolumeIterator = true
	case "GetVolSnapsError":
		mock.InducedErrors.GetVolSnapsError = true
	case "GetPrivVolumeByIDError":
		mock.InducedErrors.GetPrivVolumeByIDError = true
	case "CreatePortGroupError":
		mock.InducedErrors.CreatePortGroupError = true
	case "UpdatePortGroupError":
		mock.InducedErrors.UpdatePortGroupError = true
	case "DeletePortGroupError":
		mock.InducedErrors.DeletePortGroupError = true
	case "none":
	default:
		return fmt.Errorf("unknown errorType: %s", errorType)
	}
	return nil
}

func (c *unitContext) aValidConnection() error {
	c.reset()
	mock.Reset()
	if c.client == nil {
		err := c.iCallAuthenticateWithEndpointCredentials("", "")
		if err != nil {
			return err
		}
	}
	c.checkGoRoutines("aValidConnection")
	c.client.SetAllowedArrays([]string{})
	return nil
}

func (c *unitContext) checkGoRoutines(tag string) {
	goroutines := runtime.NumGoroutine()
	fmt.Printf("goroutines %s new %d old groutines %d\n", tag, goroutines, c.nGoRoutines)
	c.nGoRoutines = goroutines
}

func (c *unitContext) iCallAuthenticateWithEndpointCredentials(endpoint, credentials string) error {
	URL := mockServer.URL
	switch endpoint {
	case "badurl":
		URL = "https://127.0.0.99:2222"
	case "nilurl":
		URL = ""
	}
	apiVersion := strings.TrimSpace(os.Getenv("APIVersion"))
	fmt.Printf("apiVersion: %s\n", apiVersion)
	client, err := NewClientWithArgs(URL, apiVersion, "", true, false)
	if err != nil {
		c.err = err
		return nil
	}
	password := defaultPassword
	if credentials == "bad" {
		password = "xxx"
	}
	err = client.Authenticate(&ConfigConnect{
		Endpoint: endpoint,
		Username: defaultUsername,
		Password: password,
	})
	if err == nil {
		c.client = client
	}
	c.err = err
	return nil
}

func (c *unitContext) theErrorMessageContains(expected string) error {
	if expected == "none" {
		if c.err == nil {
			return nil
		}
		return fmt.Errorf("Unexpected error: %s", c.err)
	}
	// We expected an error message
	if c.err == nil {
		return fmt.Errorf("Expected error message %s but no error was recorded", expected)
	}
	if strings.Contains(c.err.Error(), expected) {
		return nil
	}
	return fmt.Errorf("Expected error message to contain: %s but the error message was: %s", expected, c.err)
}

func (c *unitContext) iCallGetSymmetrixIDList() error {
	c.symIDList, c.err = c.client.GetSymmetrixIDList()
	return nil
}

func (c *unitContext) iGetAValidSymmetrixIDListIfNoError() error {
	if c.err == nil {
		if c.symIDList == nil {
			return fmt.Errorf("SymmetrixIDList nil")
		}
		if len(c.symIDList.SymmetrixIDs) == 0 {
			return fmt.Errorf("No IDs in SymmetrixIDList")
		}
	}
	return nil
}

func (c *unitContext) iCallGetSymmetrixByID(id string) error {
	c.sym, c.err = c.client.GetSymmetrixByID(id)
	return nil
}

func (c *unitContext) iGetAValidSymmetrixObjectIfNoError() error {
	if c.err == nil {
		if c.sym == nil {
			return fmt.Errorf("Symmetrix nil")
		}
		fmt.Printf("Symmetrix: %#v", c.sym)
		if c.sym.SymmetrixID == "" || c.sym.Ucode == "" || c.sym.Model == "" || c.sym.DiskCount <= 0 {
			return fmt.Errorf("Problem with Symmetrix fields SymmetrixID Ucode Model or DiskCount")
		}
	}
	return nil
}

func (c *unitContext) iHaveVolumes(number int) error {
	for i := 1; i <= number; i++ {
		id := fmt.Sprintf("%05d", i)
		size := 7
		volumeIdentifier := "Vol" + id
		mock.AddNewVolume(id, volumeIdentifier, size, mock.DefaultStorageGroup)
		c.volIDList = append(c.volIDList, id)
		//mock.Data.VolumeIDToIdentifier[id] = fmt.Sprintf("Vol%05d", i)
		//mock.Data.VolumeIDToSGList[id] = make([]string, 0)
	}
	return nil
}

func (c *unitContext) iCallGetVolumeByID(volID string) error {
	c.vol, c.err = c.client.GetVolumeByID(symID, volID)
	return nil
}

func (c *unitContext) iGetAValidVolumeObjectIfNoError(id string) error {
	if c.err != nil {
		return nil
	}
	if c.vol.VolumeID != id {
		return fmt.Errorf("Expected volume %s but got %s", id, c.vol.VolumeID)
	}
	return nil
}

func (c *unitContext) iCallGetVolumeIDList(volumeIdentifier string) error {
	var like bool
	if strings.Contains(volumeIdentifier, "<like>") {
		volumeIdentifier = strings.TrimPrefix(volumeIdentifier, "<like>")
		like = true
	}
	c.volList, c.err = c.client.GetVolumeIDList(symID, volumeIdentifier, like)
	return nil
}

func (c *unitContext) iGetAValidVolumeIDListWithIfNoError(nvols int) error {
	if c.err != nil {
		return nil
	}
	if len(c.volList) != nvols {
		return fmt.Errorf("Expected %d volumes but got %d", nvols, len(c.volList))
	}
	return nil
}

func (c *unitContext) iExpandVolumeToSize(volumeID string, sizeStr string) error {
	if c.err != nil {
		return nil
	}

	if size, err := strconv.Atoi(sizeStr); err == nil {
		c.vol, c.err = c.client.ExpandVolume(symID, volumeID, size)
	} else {
		return err
	}

	return nil
}

func (c *unitContext) iValidateVolumeSize(volumeID string, sizeStr string) error {
	if c.err != nil {
		return nil
	}

	c.vol, c.err = c.client.GetVolumeByID(symID, volumeID)
	size, err := strconv.Atoi(sizeStr)
	if err == nil && float64(size) != c.vol.CapacityGB {
		return fmt.Errorf("Expected volume %s to be size %s, but was %d", volumeID, sizeStr, size)
	} else {
		return err
	}

	return nil
}

func (c *unitContext) iCallGetStorageGroupIDList() error {
	c.storageGroupIDList, c.err = c.client.GetStorageGroupIDList(symID)
	return nil
}

func (c *unitContext) iGetAValidStorageGroupIDListIfNoErrors() error {
	if c.err != nil {
		return nil
	}
	if len(c.storageGroupIDList.StorageGroupIDs) == 0 {
		return fmt.Errorf("Expected storage group IDs to be returned but there were none")
	}
	for _, id := range c.storageGroupIDList.StorageGroupIDs {
		fmt.Printf("StorageGroup: %s\n", id)
	}
	return nil
}

func (c *unitContext) iCallGetStorageGroup(sgID string) error {
	c.storageGroup, c.err = c.client.GetStorageGroup(symID, sgID)
	return nil
}

func (c *unitContext) iGetAValidStorageGroupIfNoErrors() error {
	if c.err != nil {
		return nil
	}
	if c.storageGroup.StorageGroupID == "" || c.storageGroup.Type == "" {
		return fmt.Errorf("Expected StorageGroup to have StorageGroupID and Type but didn't")
	}
	return nil
}

func (c *unitContext) iCallGetStoragePool(srpID string) error {
	c.storagePool, c.err = c.client.GetStoragePool(symID, srpID)
	return nil
}

func (c *unitContext) iGetAValidGetStoragePoolIfNoErrors() error {
	if c.err != nil {
		return nil
	}
	if c.storagePool.StoragePoolID == "" {
		return fmt.Errorf("Expected StoragePool to have StoragePoolID and Type but didn't")
	}
	return nil
}

func (c *unitContext) iHaveJobs(numberOfJobs int) error {
	for i := 1; i <= numberOfJobs; i++ {
		jobID := fmt.Sprintf("job%d", i)
		mock.NewMockJob(jobID, "RUNNING", "SUCCEEDED", "")
	}
	return nil
}

func (c *unitContext) iCallGetJobIDListWith(statusQuery string) error {
	c.jobIDList, c.err = c.client.GetJobIDList(symID, statusQuery)
	return nil
}

func (c *unitContext) iGetAValidJobsIDListWithIfNoErrors(numberOfEntries int) error {
	if c.err != nil {
		return nil
	}
	if len(c.jobIDList) != numberOfEntries {
		return fmt.Errorf("Expected %d jobs ids to be returned but got %d", numberOfEntries, len(c.jobIDList))
	}
	return nil
}

func (c *unitContext) iCreateAJobWithInitialStateAndFinalState(initialState, finalState string) error {
	mock.NewMockJob("myjob", initialState, finalState, "")
	return nil
}

func (c *unitContext) iCallGetJobByID() error {
	c.job, c.err = c.client.GetJobByID(symID, "myjob")
	return nil
}

func (c *unitContext) iGetAValidJobWithStateIfNoError(expectedState string) error {
	if c.err != nil {
		return nil
	}
	if c.job.Status != expectedState {
		return fmt.Errorf("Expected job state to be %s but instead it was %s", expectedState, c.job.Status)
	}
	return nil
}

func (c *unitContext) iCallWaitOnJobCompletion() error {
	c.job, c.err = c.client.WaitOnJobCompletion(symID, "myjob")
	return nil
}

func (c *unitContext) iCallCreateVolumeInStorageGroupWithNameAndSize(volumeName string, sizeInCylinders int) error {
	c.vol, c.err = c.client.CreateVolumeInStorageGroup(symID, mock.DefaultStorageGroup, volumeName, sizeInCylinders)
	return nil
}

func (c *unitContext) iGetAValidVolumeWithNameIfNoError(volumeName string) error {
	if c.err != nil {
		return nil
	}
	if c.vol.VolumeIdentifier != volumeName {
		return fmt.Errorf("Expected volume named %s but got %s", volumeName, c.vol.VolumeIdentifier)
	}
	return nil
}

func (c *unitContext) iCallCreateStorageGroupWithNameAndSrpAndSl(sgName, srp, serviceLevel string) error {
	c.storageGroup, c.err = c.client.CreateStorageGroup(symID, sgName, srp, serviceLevel, false)
	return nil
}

func (c *unitContext) iGetAValidStorageGroupWithNameIfNoError(sgName string) error {
	if c.err != nil {
		return nil
	}
	if c.storageGroup.StorageGroupID != sgName {
		return fmt.Errorf("Expected StorageGroup to have name %s", sgName)
	}
	return nil
}

func (c *unitContext) iCallDeleteStorageGroup(sgID string) error {
	c.err = c.client.DeleteStorageGroup(symID, sgID)
	return nil
}

func (c *unitContext) iCallGetStoragePoolList() error {
	c.storagePoolList, c.err = c.client.GetStoragePoolList(symID)
	return nil
}

func (c *unitContext) iGetAValidStoragePoolListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.storagePoolList == nil || len(c.storagePoolList.StoragePoolIDs) <= 0 || c.storagePoolList.StoragePoolIDs[0] != "SRP_1" {
		return fmt.Errorf("Expected StoragePoolList to have SRP_1 but it didn't")
	}
	return nil
}

func (c *unitContext) iCallRemoveVolumeFromStorageGroup() error {
	c.storageGroup, c.err = c.client.RemoveVolumesFromStorageGroup(symID, mock.DefaultStorageGroup, c.vol.VolumeID)
	return nil
}

func (c *unitContext) theVolumeIsNoLongerAMemberOfTheStorageGroupIfNoError() error {
	if c.err != nil {
		return nil
	}
	sgIDList := mock.Data.VolumeIDToSGList[c.vol.VolumeID]
	if len(sgIDList) == 0 {
		return nil
	}
	for _, sgid := range sgIDList {
		if sgid == mock.DefaultStorageGroup {
			return fmt.Errorf("Volume contained the Storage Group %s which was not expected", mock.DefaultStorageGroup)
		}
	}
	return nil
}

func (c *unitContext) iCallRenameVolumeWith(newName string) error {
	c.vol, c.err = c.client.RenameVolume(symID, c.vol.VolumeID, newName)
	return nil
}

func (c *unitContext) iCallInitiateDeallocationOfTracksFromVolume() error {
	c.job, c.err = c.client.InitiateDeallocationOfTracksFromVolume(symID, c.vol.VolumeID)
	return nil
}

func (c *unitContext) iCallDeleteVolume() error {
	c.err = c.client.DeleteVolume(symID, c.vol.VolumeID)
	return nil
}

func (c *unitContext) iHaveAMaskingView(maskingViewID string) error {
	sgID := maskingViewID + "-sg"
	pgID := maskingViewID + "-pg"
	hostID := maskingViewID + "-host"
	localMaskingView := &uMV{
		maskingViewID:  maskingViewID,
		hostID:         hostID,
		hostGroupID:    "",
		storageGroupID: sgID,
		portGroupID:    pgID,
	}
	initiators := []string{testInitiatorIQN}
	mock.AddInitiator(testInitiator, testInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	mock.AddHost(hostID, "iSCSI", initiators)
	mock.AddStorageGroup(sgID, "SRP_1", "Diamond")
	mock.AddMaskingView(maskingViewID, sgID, hostID, pgID)
	c.uMaskingView = localMaskingView
	return nil
}

func (c *unitContext) iCallGetMaskingViewList() error {
	c.maskingViewList, c.err = c.client.GetMaskingViewList(symID)
	return nil
}

func (c *unitContext) iGetAValidMaskingViewListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.maskingViewList == nil || len(c.maskingViewList.MaskingViewIDs) == 0 {
		return fmt.Errorf("Expected item in MaskingViewList but got none")
	}
	found := false
	for _, id := range c.maskingViewList.MaskingViewIDs {
		fmt.Printf("MaskingView: %s\n", id)
		if c.uMaskingView.maskingViewID == id {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Expected to find %s in MaskingViewList but didn't", c.uMaskingView.maskingViewID)
	}
	return nil
}

func (c *unitContext) iCallGetMaskingViewByID(mvID string) error {
	c.maskingView, c.err = c.client.GetMaskingViewByID(symID, mvID)
	return nil
}

func (c *unitContext) iGetAValidMaskingViewIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.maskingView == nil {
		return fmt.Errorf("Expecting a masking view but received none")
	}
	if c.hostGroupID == "" {
		if c.maskingView.HostID != c.uMaskingView.hostID {
			return fmt.Errorf("Expecting host %s but got %s", c.uMaskingView.hostID, c.maskingView.HostID)
		}
	} else {
		if c.maskingView.HostID != c.uMaskingView.hostGroupID {
			return fmt.Errorf("Expecting hostgroup %s but got %s", c.uMaskingView.hostGroupID, c.maskingView.HostID)
		}
	}
	if c.maskingView.PortGroupID != c.uMaskingView.portGroupID {
		return fmt.Errorf("Expecting portgroup %s but got %s", c.uMaskingView.portGroupID, c.maskingView.PortGroupID)
	}
	if c.maskingView.StorageGroupID != c.uMaskingView.storageGroupID {
		return fmt.Errorf("Expecting storagegroup %s but got %s", c.uMaskingView.storageGroupID, c.maskingView.StorageGroupID)
	}
	return nil
}

func (c *unitContext) iCallDeleteMaskingView() error {
	c.err = c.client.DeleteMaskingView(symID, c.uMaskingView.maskingViewID)
	return nil
}

func (c *unitContext) iHaveAPortGroup() error {
	mock.AddPortGroup(testPortGroup, "ISCSI", []string{"SE-1E:000"})
	return nil
}

func (c *unitContext) iCallGetPortGroupList() error {
	c.portGroupList, c.err = c.client.GetPortGroupList(symID, "")
	return nil
}

func (c *unitContext) iGetAValidPortGroupListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.portGroupList == nil || len(c.portGroupList.PortGroupIDs) == 0 {
		return fmt.Errorf("Expected item in PortGroupList but got none")
	}
	return nil
}

func (c *unitContext) iCallGetPortGroupByID() error {
	c.portGroup, c.err = c.client.GetPortGroupByID(symID, testPortGroup)
	return nil
}

func (c *unitContext) iCallCreatePortGroup(groupName string, strSliceOfPorts string) error {
	if c.err != nil {
		return nil
	}

	initialPorts := convertStringSliceOfPortsToPortKeys(strSliceOfPorts)
	c.portGroup, c.err = c.client.CreatePortGroup(symID, groupName, initialPorts)
	return nil
}

func (c *unitContext) iCallUpdatePortGroup(groupName string, strUpdatePorts string) error {
	if c.err != nil {
		return nil
	}

	updatedPorts := convertStringSliceOfPortsToPortKeys(strUpdatePorts)
	c.portGroup, c.err = c.client.UpdatePortGroup(symID, groupName, updatedPorts)
	return nil
}

func (c *unitContext) iExpectedThesePortsInPortGroup(strSliceOfPorts string) error {
	if c.err != nil {
		return nil
	}

	expectedPorts := convertStringSliceOfPortsToPortKeys(strSliceOfPorts)
	if c.portGroup == nil {
		return errors.New("could not find any portGroup. Make sure test was set up with a PortGroup")
	}

	expectedPortsLen := len(expectedPorts)
	portGroupLen := len(c.portGroup.SymmetrixPortKey)
	if expectedPortsLen != portGroupLen {
		return fmt.Errorf("expected number of ports does not match. Expected %d, but portGroup %s has %d", expectedPortsLen, c.portGroup.PortGroupID, portGroupLen)
	}

	portKeySlice := make([]string, 0)
	portsInPortGroup := make(map[string]bool)
	for _, its := range c.portGroup.SymmetrixPortKey {
		thisKey := fmt.Sprintf("%s:%s", its.DirectorID, its.PortID)
		if !portsInPortGroup[thisKey] {
			portsInPortGroup[thisKey] = true
			portKeySlice = append(portKeySlice, thisKey)
		}
	}

	for _, its := range expectedPorts {
		thisKey := fmt.Sprintf("%s:%s", its.DirectorID, its.PortID)
		if !portsInPortGroup[thisKey] {
			return fmt.Errorf("list of ports in PortGroup do not match expected list. Expected %s, but got %s", strSliceOfPorts, strings.Join(portKeySlice, ","))
		}
	}

	return nil
}

func (c *unitContext) iCallDeletePortGroup(groupName string) error {
	if c.err != nil {
		return nil
	}

	c.err = c.client.DeletePortGroup(symID, groupName)

	return nil
}

func (c *unitContext) thePortGroupShouldNotExist(groupName string) error {
	if c.err != nil {
		return nil
	}

	c.portGroupList, c.err = c.client.GetPortGroupList(symID, "")
	for _, id := range c.portGroupList.PortGroupIDs {
		if id == groupName {
			return fmt.Errorf("PortGroup %s was not expected, but is in PortGroup list", groupName)
		}
	}
	return nil
}

func (c *unitContext) iGetPortGroupIfNoError(groupName string) error {
	if c.err != nil {
		return nil
	}

	if c.portGroup.PortGroupID != groupName {
		return fmt.Errorf("Expected to get Port Group %s, but received %s",
			c.portGroup.PortGroupID, groupName)
	}
	return nil
}

func (c *unitContext) iGetAValidPortGroupIfNoError() error {
	if c.err != nil {
		return nil
	}

	if c.portGroup.PortGroupID != testPortGroup {
		return fmt.Errorf("Expected to get Port Group %s, but received %s",
			c.portGroup.PortGroupID, testPortGroup)
	}
	return nil
}

func (c *unitContext) iHaveAISCSIHost(hostName string) error {
	initiators := []string{testInitiatorIQN}
	mock.AddInitiator(testInitiator, testInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	c.hostID = hostName
	c.host, c.err = mock.AddHost(c.hostID, "iSCSI", initiators)
	return nil
}

func (c *unitContext) iHaveAFCHost(hostName string) error {
	initiators := []string{testFCInitiatorWWN}
	mock.AddInitiator(testFCInitiator, testFCInitiatorWWN, "Fibre", []string{"FA-1D:4"}, "")
	c.hostID = hostName
	c.host, c.err = mock.AddHost(c.hostID, "Fibre", initiators)
	return nil
}

func (c *unitContext) iCallGetHostList() error {
	c.hostList, c.err = c.client.GetHostList(symID)
	return nil
}

func (c *unitContext) iGetAValidHostListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.hostList == nil || len(c.hostList.HostIDs) == 0 {
		return fmt.Errorf("Expected item in HostList but got none")
	}
	fmt.Println(c.hostList)
	return nil
}

func (c *unitContext) iCallGetHostByID(hostID string) error {
	c.host, c.err = c.client.GetHostByID(symID, hostID)
	return nil
}

func (c *unitContext) iGetAValidHostIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.host.HostID != c.hostID {
		return fmt.Errorf("Expected to get Host %s, but received %s",
			c.host.HostID, c.hostID)
	}
	return nil
}

func (c *unitContext) iCallCreateHost(hostName string) error {
	initiatorList := make([]string, 1)
	c.hostID = hostName
	initiatorList[0] = testInitiatorIQN
	mock.AddInitiator(testInitiator, testInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	c.host, c.err = c.client.CreateHost(symID, hostName, initiatorList, nil)
	return nil
}

func (c *unitContext) iCallUpdateHost() error {
	initiatorList := make([]string, 1)
	initiatorList[0] = testUpdateInitiatorIQN
	mock.AddInitiator(testUpdateInitiator, testUpdateInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	c.host, c.err = c.client.UpdateHostInitiators(symID, c.host, initiatorList)
	return nil
}

func (c *unitContext) iCallDeleteHost(hostName string) error {
	c.err = c.client.DeleteHost(symID, hostName)
	return nil
}

func (c *unitContext) iHaveAInitiator() error {
	return nil
}

func (c *unitContext) iCallGetInitiatorList() error {
	c.initiatorList, c.err = c.client.GetInitiatorList(symID, "", false, false)
	return nil
}

func (c *unitContext) iCallGetInitiatorListWithFilters() error {
	c.initiatorList, c.err = c.client.GetInitiatorList(symID, testInitiatorIQN, true, true)
	return nil
}

func (c *unitContext) iGetAValidInitiatorListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.initiatorList == nil || len(c.initiatorList.InitiatorIDs) == 0 {
		return fmt.Errorf("Expected item in InitiatorList but got none")
	}
	return nil
}

func (c *unitContext) iCallGetInitiatorByID() error {
	mock.AddInitiator(testInitiator, testInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	c.initiator, c.err = c.client.GetInitiatorByID(symID, testInitiator)
	return nil
}

func (c *unitContext) iGetAValidInitiatorIfNoError() error {
	if c.err != nil {
		return nil
	}

	if c.initiator.InitiatorID != testInitiatorIQN {
		return fmt.Errorf("Expected to get initiator %s, but received %s",
			c.initiator.InitiatorID, testInitiator)
	}
	return nil
}

func (c *unitContext) iHaveAStorageGroup(sgID string) error {
	c.sgID = sgID
	mock.AddStorageGroup(sgID, "SRP_1", "Diamond")
	return nil
}

func (c *unitContext) iCallCreateMaskingViewWithHost(mvID string) error {
	localMaskingView := &uMV{
		maskingViewID:  mvID,
		hostID:         c.hostID,
		hostGroupID:    "",
		storageGroupID: c.sgID,
		portGroupID:    testPortGroup,
	}
	c.uMaskingView = localMaskingView
	c.maskingView, c.err = c.client.CreateMaskingView(symID, mvID, c.sgID, c.hostID, true, testPortGroup)
	return nil
}

func (c *unitContext) iCallCreateMaskingViewWithHostGroup(mvID string) error {
	localMaskingView := &uMV{
		maskingViewID:  mvID,
		hostID:         "",
		hostGroupID:    c.hostGroupID,
		storageGroupID: c.sgID,
		portGroupID:    testPortGroup,
	}
	c.uMaskingView = localMaskingView
	c.maskingView, c.err = c.client.CreateMaskingView(symID, mvID, c.sgID, c.hostGroupID, false, testPortGroup)
	return nil
}

func (c *unitContext) iHaveAHostGroup(hostGroupID string) error {
	// Create a host instead of host group
	c.hostGroupID = hostGroupID
	initiators := []string{testInitiatorIQN}
	mock.AddInitiator(testInitiator, testInitiatorIQN, "GigE", []string{"SE-1E:000"}, "")
	mock.AddHost(hostGroupID, "iSCSI", initiators)
	return nil
}

func (c *unitContext) iCallAddVolumesToStorageGroup(sgID string) error {
	c.err = c.client.AddVolumesToStorageGroup(symID, sgID, c.volIDList...)
	return nil
}

func (c *unitContext) thenTheVolumesArePartOfStorageGroupIfNoError() error {
	if c.err != nil {
		return nil
	}
	sgList := mock.Data.VolumeIDToSGList["00001"]
	fmt.Printf("%v", sgList)
	for volumeID := range mock.Data.VolumeIDToIdentifier {
		fmt.Println(volumeID)
		sgList := mock.Data.VolumeIDToSGList[volumeID]
		fmt.Printf("%v", sgList)
		volumeFound := false
		for _, sg := range sgList {
			if sg == mock.DefaultStorageGroup {
				volumeFound = true
				break
			}
		}
		if !volumeFound {
			return fmt.Errorf("Couldn't find volume in storage group")
		}
	}
	return nil
}

func (c *unitContext) iCallGetListOfTargetAddresses() error {
	c.addressList, c.err = c.client.GetListOfTargetAddresses(symID)
	return nil
}

func (c *unitContext) iRecieveIPAddresses(count int) error {
	if len(c.addressList) != count {
		return fmt.Errorf("Expected to get %d addresses but recieved %d", count, len(c.addressList))
	}
	return nil
}

func (c *unitContext) iHaveAWhitelistOf(whitelist string) error {
	// turn the whitelist string into a slice
	results := convertStringToSlice(whitelist)

	// set the whitelist
	c.client.SetAllowedArrays(results)
	return nil
}

func (c *unitContext) itContainsArrays(count int) error {
	allowed := c.client.GetAllowedArrays()
	if len(allowed) != count {
		return fmt.Errorf("Received the wrong number of arrays in the whitelist. Expected %d but have a whitelist of %v", count, allowed)
	}
	return nil
}

func (c *unitContext) shouldInclude(include string) error {
	// turn the list of specified arrays into a slice
	results := convertStringToSlice(include)

	// make sure each one is in the whitelist
	for _, a := range results {
		if ok, _ := c.client.IsAllowedArray(a); ok == false {
			return fmt.Errorf("Expected array (%s) to be in the whitelist but it was not found", a)
		}
	}
	return nil
}

func (c *unitContext) shouldNotInclude(exclude string) error {
	// turn the list of specified arrays into a slice
	results := convertStringToSlice(exclude)

	// make sure each one is not in the whitelist
	for _, a := range results {
		if ok, _ := c.client.IsAllowedArray(a); ok == true {
			return fmt.Errorf("Expected array (%s) to not be in the whitelist but it was", a)
		}
	}
	return nil
}

func (c *unitContext) iGetAValidSymmetrixIDListThatContainsAndDoesNotContains(included string, excluded string) error {
	includedArrays := convertStringToSlice(included)
	// make sure all the included arrays exist in the response
	for _, array := range includedArrays {
		found := false
		for _, expectedArray := range c.symIDList.SymmetrixIDs {
			if array == expectedArray {
				found = true
			}
		}
		if found == false {
			return fmt.Errorf("Expected array %s to be included in %v, but it was not", array, c.symIDList.SymmetrixIDs)
		}
	}

	excludedArrays := convertStringToSlice(excluded)
	// make sure all the excluded arrays do NOT exist in the response
	for _, array := range excludedArrays {
		found := false
		for _, expectedArray := range c.symIDList.SymmetrixIDs {
			if array == expectedArray {
				found = true
			}
		}
		if found != false {
			return fmt.Errorf("Expected array %s to be excluded in %v, but it was not", array, c.symIDList.SymmetrixIDs)
		}
	}

	return nil
}

func convertStringToSlice(input string) []string {
	results := make([]string, 0)
	st := strings.Split(input, ",")
	for i := range st {
		t := strings.TrimSpace(st[i])
		if t != "" {
			results = append(results, t)
		}
	}
	return results
}

func (c *unitContext) iExcuteTheCapabilitiesOnTheSymmetrixArray() error {
	c.symRepCapibilities, c.err = c.client.GetReplicationCapabilities()
	return nil
}

func (c *unitContext) iCallGetSnapVolumeListWithAnd(queryKey, queryValue string) error {
	if queryKey != "" {
		if queryValue == "true" {
			c.symVolumeList, c.err = c.client.GetSnapVolumeList(symID, types.QueryParams{
				queryKey: true})
		}
	} else {
		c.symVolumeList, c.err = c.client.GetSnapVolumeList(symID, nil)
	}
	return nil
}

func (c *unitContext) iShouldGetListOfVolumesHavingSnapshots() error {
	if c.err != nil {
		return nil
	}
	if len(c.symVolumeList.Name) == 0 {
		return fmt.Errorf("No Volumes with Snapshot found")
	}
	return nil
}

func (c *unitContext) iCallGetVolumeSnapInfoWithVolume(volID string) error {
	c.volSnapList, c.err = c.client.GetVolumeSnapInfo(symID, volID)
	return nil
}

func (c *unitContext) iShouldGetAListOfSnapshotsIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.volSnapList == nil {
		return fmt.Errorf("There is no Snapshot for provided source volume")
	}
	return nil
}

func (c *unitContext) iCallCreateSnapshotWithAndSnapshotOnIt(volIds, snapID string) error {
	c.sourceVolumeList = c.createVolumeList(volIds)
	c.err = c.client.CreateSnapshot(symID, snapID, c.sourceVolumeList, 0)
	return nil
}

func (c *unitContext) iGetAValidSnapshotObjectIfNoError() error {
	if c.err != nil {
		return nil
	}
	for i := range c.sourceVolumeList {
		sourceVol := c.sourceVolumeList[i]
		if mock.Data.VolIDToSnapshots[sourceVol.Name] == nil {
			return fmt.Errorf("The snaphshot does not exist for source volume %s", sourceVol.Name)
		}
	}
	return nil
}

func (c *unitContext) iCallGetSnapshotInfoWithAndSnapshotNameOnIt(volID, SnapID string) error {
	c.volumeSnapshot, c.err = c.client.GetSnapshotInfo(symID, volID, SnapID)
	return nil
}

func (c *unitContext) iShouldGetTheSnapshotDetailsIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.volumeSnapshot == nil {
		return fmt.Errorf("The Snapshot does not exist")
	}
	return nil
}

func (c *unitContext) iCallGetSnapshotGenerationsWithAndSnapshotOnIt(volID, SnapID string) error {
	c.volSnapGenerationList, c.err = c.client.GetSnapshotGenerations(symID, volID, SnapID)
	return nil
}

func (c *unitContext) iShouldGetTheGenerationListIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.volSnapGenerationList == nil {
		return fmt.Errorf("The Generation List for the snaphshot does not exist")
	}
	return nil
}

func (c *unitContext) iCallGetSnapshotGenerationWithSnapshotAndOnIt(volID string, snapID string, genID int64) error {
	c.volSnapGenerationInfo, c.err = c.client.GetSnapshotGenerationInfo(symID, volID, snapID, genID)
	return nil
}

func (c *unitContext) iShouldGetAGenerationInfoIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.volSnapGenerationInfo == nil {
		return fmt.Errorf("The Generation for the snaphshot does not exist")
	}
	return nil
}

func (c *unitContext) iCallModifySnapshotWithAnd(sourceVols, targetVols, SnapID, newSnapID string, genID int64, action string) error {
	sourceVolumeList := c.createVolumeList(sourceVols)
	targetVolumeList := c.createVolumeList(targetVols)
	c.err = c.client.ModifySnapshot(symID, sourceVolumeList, targetVolumeList, SnapID, action, newSnapID, genID)

	return nil
}

func (c *unitContext) iCallDeleteSnapshotWithSnapshotAndOnIt(sourceVols, SnapID string, genID int64) error {
	sourceVolumeList := c.createVolumeList(sourceVols)
	c.err = c.client.DeleteSnapshot(symID, SnapID, sourceVolumeList, genID)
	return nil
}

func (c *unitContext) iCallGetPrivVolumeByIDWith(volID string) error {
	c.volResultPrivate, c.err = c.client.GetPrivVolumeByID(symID, volID)
	return nil
}

func (c *unitContext) iShouldGetAPrivateVolumeInformationIfNoError() error {
	if c.err != nil {
		return nil
	}
	if c.volResultPrivate == nil {
		return fmt.Errorf("The private information for the volume does not exist")
	}
	return nil
}

func (c *unitContext) iShouldGetAValidResponseIfNoError() error {
	if c.err != nil {
		return nil
	}
	return nil
}

func (c *unitContext) thereShouldBeNoErrors() error {
	return c.err
}

// createVolumeList will extract all the volumes and will return a list of type VolumeList
func (c *unitContext) createVolumeList(volIds string) []types.VolumeList {
	var VolumeList []types.VolumeList
	volNames := strings.Split(volIds, ",")
	for i := 0; i < len(volNames); i++ {
		VolumeList = append(VolumeList, types.VolumeList{Name: volNames[i]})
	}
	return VolumeList
}

// convertStringSliceOfPortsToPortKeys - Given a comma delimited string of ports in
// this format "<DIRECTOR>:<PORT>", produce a slice of types.PortKey values
func convertStringSliceOfPortsToPortKeys(strListOfPorts string) []types.PortKey {
	initialPorts := make([]types.PortKey, 0)
	for _, it := range convertStringToSlice(strListOfPorts) {
		dirAndPort := strings.Split(it, ":")
		port := types.PortKey{
			DirectorID: dirAndPort[0],
			PortID:     dirAndPort[1],
		}
		initialPorts = append(initialPorts, port)
	}
	return initialPorts
}

func UnitTestContext(s *godog.Suite) {
	c := &unitContext{}
	s.Step(`^I induce error "([^"]*)"$`, c.iInduceError)
	s.Step(`^I call authenticate with endpoint "([^"]*)" credentials "([^"]*)"$`, c.iCallAuthenticateWithEndpointCredentials)
	s.Step(`^the error message contains "([^"]*)"$`, c.theErrorMessageContains)
	s.Step(`^a valid connection$`, c.aValidConnection)
	s.Step(`^I call GetSymmetrixIDList$`, c.iCallGetSymmetrixIDList)
	s.Step(`^I get a valid Symmetrix ID List if no error$`, c.iGetAValidSymmetrixIDListIfNoError)
	s.Step(`^I call GetSymmetrixByID "([^"]*)"$`, c.iCallGetSymmetrixByID)
	s.Step(`^I get a valid Symmetrix Object if no error$`, c.iGetAValidSymmetrixObjectIfNoError)
	s.Step(`^I have (\d+) volumes$`, c.iHaveVolumes)
	s.Step(`^I call GetVolumeByID "([^"]*)"$`, c.iCallGetVolumeByID)
	s.Step(`^I get a valid Volume Object "([^"]*)" if no error$`, c.iGetAValidVolumeObjectIfNoError)
	s.Step(`^I call GetVolumeIDList "([^"]*)"$`, c.iCallGetVolumeIDList)
	s.Step(`^I get a valid VolumeIDList with (\d+) if no error$`, c.iGetAValidVolumeIDListWithIfNoError)
	s.Step(`^I call GetStorageGroupIDList$`, c.iCallGetStorageGroupIDList)
	s.Step(`^I get a valid StorageGroupIDList if no errors$`, c.iGetAValidStorageGroupIDListIfNoErrors)
	s.Step(`^I call GetStorageGroup "([^"]*)"$`, c.iCallGetStorageGroup)
	s.Step(`^I have a StorageGroup "([^"]*)"$`, c.iHaveAStorageGroup)
	s.Step(`^I get a valid StorageGroup if no errors$`, c.iGetAValidStorageGroupIfNoErrors)
	s.Step(`^I have (\d+) jobs$`, c.iHaveJobs)
	s.Step(`^I call GetJobIDList with "([^"]*)"$`, c.iCallGetJobIDListWith)
	s.Step(`^I get a valid JobsIDList with (\d+) if no errors$`, c.iGetAValidJobsIDListWithIfNoErrors)
	s.Step(`^I create a job with initial state "([^"]*)" and final state "([^"]*)"$`, c.iCreateAJobWithInitialStateAndFinalState)
	s.Step(`^I call GetJobByID$`, c.iCallGetJobByID)
	s.Step(`^I get a valid Job with state "([^"]*)" if no error$`, c.iGetAValidJobWithStateIfNoError)
	s.Step(`^I call WaitOnJobCompletion$`, c.iCallWaitOnJobCompletion)
	// Volumes
	s.Step(`^I call CreateVolumeInStorageGroup with name "([^"]*)" and size (\d+)$`, c.iCallCreateVolumeInStorageGroupWithNameAndSize)
	s.Step(`^I get a valid Volume with name "([^"]*)" if no error$`, c.iGetAValidVolumeWithNameIfNoError)
	s.Step(`^I call CreateStorageGroup with name "([^"]*)" and srp "([^"]*)" and sl "([^"]*)"$`, c.iCallCreateStorageGroupWithNameAndSrpAndSl)
	s.Step(`^I call DeleteStorageGroup "([^"]*)"$`, c.iCallDeleteStorageGroup)
	s.Step(`^I get a valid StorageGroup with name "([^"]*)" if no error$`, c.iGetAValidStorageGroupWithNameIfNoError)
	s.Step(`^I call GetStoragePoolList$`, c.iCallGetStoragePoolList)
	s.Step(`^I get a valid StoragePoolList if no error$`, c.iGetAValidStoragePoolListIfNoError)
	s.Step(`^I call RemoveVolumeFromStorageGroup$`, c.iCallRemoveVolumeFromStorageGroup)
	s.Step(`^the volume is no longer a member of the Storage Group if no error$`, c.theVolumeIsNoLongerAMemberOfTheStorageGroupIfNoError)
	s.Step(`^I call RenameVolume with "([^"]*)"$`, c.iCallRenameVolumeWith)
	s.Step(`^I call InitiateDeallocationOfTracksFromVolume$`, c.iCallInitiateDeallocationOfTracksFromVolume)
	s.Step(`^I call DeleteVolume$`, c.iCallDeleteVolume)
	s.Step(`^I expand volume "([^"]*)" to "([^"]*)" in GB$`, c.iExpandVolumeToSize)
	s.Step(`^I validate that volume "([^"]*)" has has size "([^"]*)" in GB$`, c.iValidateVolumeSize)
	// Masking View
	s.Step(`^I have a MaskingView "([^"]*)"$`, c.iHaveAMaskingView)
	s.Step(`^I call GetMaskingViewList$`, c.iCallGetMaskingViewList)
	s.Step(`^I get a valid MaskingViewList if no error$`, c.iGetAValidMaskingViewListIfNoError)
	s.Step(`^I call GetMaskingViewByID "([^"]*)"$`, c.iCallGetMaskingViewByID)
	s.Step(`^I get a valid MaskingView if no error$`, c.iGetAValidMaskingViewIfNoError)
	s.Step(`^I call CreateMaskingViewWithHost "([^"]*)"$`, c.iCallCreateMaskingViewWithHost)
	s.Step(`^I call CreateMaskingViewWithHostGroup "([^"]*)"$`, c.iCallCreateMaskingViewWithHostGroup)
	s.Step(`^I call DeleteMaskingView$`, c.iCallDeleteMaskingView)
	// Port Group
	s.Step(`^I have a PortGroup$`, c.iHaveAPortGroup)
	s.Step(`^I call GetPortGroupList$`, c.iCallGetPortGroupList)
	s.Step(`^I get a valid PortGroupList if no error$`, c.iGetAValidPortGroupListIfNoError)
	s.Step(`^I call GetPortGroupByID$`, c.iCallGetPortGroupByID)
	s.Step(`^I get a valid PortGroup if no error$`, c.iGetAValidPortGroupIfNoError)
	s.Step(`^I get PortGroup "([^"]*)" if no error$`, c.iGetPortGroupIfNoError)
	s.Step(`^I call CreatePortGroup "([^"]*)" with ports "([^"]*)"$`, c.iCallCreatePortGroup)
	s.Step(`^I call UpdatePortGroup "([^"]*)" with ports "([^"]*)"$`, c.iCallUpdatePortGroup)
	s.Step(`^I call DeletePortGroup "([^"]*)"$`, c.iCallDeletePortGroup)
	s.Step(`^I expect PortGroup to have these ports "([^"]*)"$`, c.iExpectedThesePortsInPortGroup)
	s.Step(`^the PortGroup "([^"]*)" should not exist`, c.thePortGroupShouldNotExist)
	// Host
	s.Step(`^I have a FC Host "([^"]*)"$`, c.iHaveAFCHost)
	s.Step(`^I have a ISCSI Host "([^"]*)"$`, c.iHaveAISCSIHost)
	s.Step(`^I call GetHostList$`, c.iCallGetHostList)
	s.Step(`^I get a valid HostList if no error$`, c.iGetAValidHostListIfNoError)
	s.Step(`^I call GetHostByID "([^"]*)"$`, c.iCallGetHostByID)
	s.Step(`^I get a valid Host if no error$`, c.iGetAValidHostIfNoError)
	// Initiator
	s.Step(`^I have a Initiator$`, c.iHaveAInitiator)
	s.Step(`^I call GetInitiatorList$`, c.iCallGetInitiatorList)
	s.Step(`^I call GetInitiatorList with filters$`, c.iCallGetInitiatorListWithFilters)
	s.Step(`^I get a valid InitiatorList if no error$`, c.iGetAValidInitiatorListIfNoError)
	s.Step(`^I call GetInitiatorByID$`, c.iCallGetInitiatorByID)
	s.Step(`^I get a valid Initiator if no error$`, c.iGetAValidInitiatorIfNoError)
	// HostGroup
	s.Step(`^I have a HostGroup "([^"]*)"$`, c.iHaveAHostGroup)
	s.Step(`^I call CreateHost "([^"]*)"$`, c.iCallCreateHost)
	s.Step(`^I call DeleteHost "([^"]*)"$`, c.iCallDeleteHost)
	s.Step(`^I call AddVolumesToStorageGroup "([^"]*)"$`, c.iCallAddVolumesToStorageGroup)
	s.Step(`^then the Volumes are part of StorageGroup if no error$`, c.thenTheVolumesArePartOfStorageGroupIfNoError)
	s.Step(`^I call UpdateHost$`, c.iCallUpdateHost)
	// GetListOftargetAddresses
	s.Step(`^I call GetListOfTargetAddresses$`, c.iCallGetListOfTargetAddresses)
	s.Step(`^I recieve (\d+) IP addresses$`, c.iRecieveIPAddresses)
	s.Step(`^I call GetStoragePool "([^"]*)"$`, c.iCallGetStoragePool)
	s.Step(`^I get a valid GetStoragePool if no errors$`, c.iGetAValidGetStoragePoolIfNoErrors)
	// Whitelists
	s.Step(`^I have a whitelist of "([^"]*)"$`, c.iHaveAWhitelistOf)
	s.Step(`^it contains (\d+) arrays$`, c.itContainsArrays)
	s.Step(`^should include "([^"]*)"$`, c.shouldInclude)
	s.Step(`^should not include "([^"]*)"$`, c.shouldNotInclude)

	s.Step(`^I get a valid Symmetrix ID List that contains "([^"]*)" and does not contains "([^"]*)"$`, c.iGetAValidSymmetrixIDListThatContainsAndDoesNotContains)

	//Snapshot
	s.Step(`^I excute the capabilities on the symmetrix array$`, c.iExcuteTheCapabilitiesOnTheSymmetrixArray)
	s.Step(`^I call GetSnapVolumeList with "([^"]*)" and "([^"]*)"$`, c.iCallGetSnapVolumeListWithAnd)
	s.Step(`^I should get a list of volumes having snapshots if no error$`, c.iShouldGetListOfVolumesHavingSnapshots)
	s.Step(`^I call GetVolumeSnapInfo with volume "([^"]*)"$`, c.iCallGetVolumeSnapInfoWithVolume)
	s.Step(`^I should get a list of snapshots if no error$`, c.iShouldGetAListOfSnapshotsIfNoError)
	s.Step(`^I call CreateSnapshot with "([^"]*)" and snapshot "([^"]*)" on it$`, c.iCallCreateSnapshotWithAndSnapshotOnIt)
	s.Step(`^I get a valid Snapshot object if no error$`, c.iGetAValidSnapshotObjectIfNoError)
	s.Step(`^I call GetSnapshotInfo with "([^"]*)" and snapshot "([^"]*)" on it$`, c.iCallGetSnapshotInfoWithAndSnapshotNameOnIt)
	s.Step(`^I should get the snapshot details if no error$`, c.iShouldGetTheSnapshotDetailsIfNoError)
	s.Step(`^I call GetSnapshotGenerations with "([^"]*)" and snapshot "([^"]*)" on it$`, c.iCallGetSnapshotGenerationsWithAndSnapshotOnIt)
	s.Step(`^I should get the generation list if no error$`, c.iShouldGetTheGenerationListIfNoError)
	s.Step(`^I call GetSnapshotGeneration with "([^"]*)", snapshot "([^"]*)" and (\d+) on it$`, c.iCallGetSnapshotGenerationWithSnapshotAndOnIt)
	s.Step(`^I should get a generation Info if no error$`, c.iShouldGetAGenerationInfoIfNoError)
	s.Step(`^I call ModifySnapshot with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", (\d+) and "([^"]*)"$`, c.iCallModifySnapshotWithAnd)
	s.Step(`^I should get a valid response if no error$`, c.iShouldGetAValidResponseIfNoError)
	s.Step(`^I call DeleteSnapshot with "([^"]*)", snapshot "([^"]*)" and (\d+)  on it$`, c.iCallDeleteSnapshotWithSnapshotAndOnIt)
	s.Step(`^I call GetPrivVolumeByID with "([^"]*)"$`, c.iCallGetPrivVolumeByIDWith)
	s.Step(`^I should get a private volume information if no error$`, c.iShouldGetAPrivateVolumeInformationIfNoError)

	s.Step(`^there should be no errors$`, c.thereShouldBeNoErrors)
}
