/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"fmt"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	cspcv1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// BlockDeviceReplacement contains old and new CSPC to validate for block device replacement
type BlockDeviceReplacement struct {
	// OldCSPC is the persisted CSPC in etcd.
	OldCSPC *apis.CStorPoolCluster
	// NewCSPC is the CSPC after it has been modified but yet not persisted to etcd.
	NewCSPC *apis.CStorPoolCluster
}

type BlockDeviceReplacementValidation struct {
	BlockDeviceReplacement BDReplacementInterface
}

// NewBlockDeviceReplacementObject returns an empty BlockDeviceReplacement object.
func NewBlockDeviceReplacementObject() *BlockDeviceReplacement {
	return &BlockDeviceReplacement{
		OldCSPC: &apis.CStorPoolCluster{},
		NewCSPC: &apis.CStorPoolCluster{},
	}
}

// WithOldCSPC sets the old persisted CSPC into the BlockDeviceReplacement object.
func (bdr *BlockDeviceReplacement) WithOldCSPC(oldCSPC *apis.CStorPoolCluster) *BlockDeviceReplacement {
	bdr.OldCSPC = oldCSPC
	return bdr
}

// WithNewCSPC sets the new CSPC as a result of CSPC modification which is not yet persisted,
// into the BlockDeviceReplacement object
func (bdr *BlockDeviceReplacement) WithNewCSPC(newCSPC *apis.CStorPoolCluster) *BlockDeviceReplacement {
	bdr.NewCSPC = newCSPC
	return bdr
}

type BDReplacementInterface interface {
	IsBDReplacementValid(newRG apis.RaidGroup, oldRG apis.RaidGroup) (bool, string)
}

// NewBlockDeviceReplacement returns a BlockDeviceReplacement interface.
func NewBlockDeviceReplacement(bdr *BlockDeviceReplacement) *BlockDeviceReplacementValidation {
	return &BlockDeviceReplacementValidation{BlockDeviceReplacement: bdr}
}

// ValidateForBDReplacementCase validates the changes in CSPC for block device replacement in a raid group only if the
// update/edit of CSPC can trigger a block device replacement.
func ValidateForBDReplacementCase(cspcNew, cspcOld *apis.CStorPoolCluster, bdrv *BlockDeviceReplacementValidation) (bool, string) {
	for i, newPool := range cspcNew.Spec.Pools {
		for j, newRaidGroup := range newPool.RaidGroups {
			if IsBlockDeviceReplacementCase(newRaidGroup, cspcOld.Spec.Pools[i].RaidGroups[j]) {
				if ok, msg := bdrv.BlockDeviceReplacement.IsBDReplacementValid(newRaidGroup, cspcOld.Spec.Pools[i].RaidGroups[j]); !ok {
					return false, msg
				}
			}
		}
	}
	return true, ""
}

// IsBlockDeviceReplacementCase returns true if the edit/update of CSPC can trigger a blockdevice
// replacement.
func IsBlockDeviceReplacementCase(newRaidGroup, oldRaidGroup apis.RaidGroup) bool {
	count := GetNumberOfDiskReplaced(newRaidGroup, oldRaidGroup)
	if count >= 1 {
		return true
	}
	return false
}

// IsMoreThanOneDiskReplaced returns true if more than one disk is replaced in the same raid group.
func GetNumberOfDiskReplaced(newRG apis.RaidGroup, oldRG apis.RaidGroup) int {
	var count int
	oldBlockDevicesMap := make(map[string]bool)
	for _, bdOld := range oldRG.BlockDevices {
		oldBlockDevicesMap[bdOld.BlockDeviceName] = true
	}
	for _, newBD := range newRG.BlockDevices {
		if !oldBlockDevicesMap[newBD.BlockDeviceName] {
			count++
		}
	}
	return count
}

// IsBDReplacementValid validates for BD replacement.
func (bdr *BlockDeviceReplacement) IsBDReplacementValid(newRG apis.RaidGroup, oldRG apis.RaidGroup) (bool, string) {

	// Not more than 1 bd should be replaced in a raid group.
	if IsMoreThanOneDiskReplaced(newRG, oldRG) {
		return false, "cannot replace more than one blockdevice in a raid group"
	}

	// The incoming BD for replacement should not pe present in the current CSPC.
	if bdr.IsNewBDPresentOnCurrentCSPC(newRG, oldRG) {
		return false, "the new blockdevice intended to use for replacement is already a part of the current cspc"
	}

	// No background replacement should be going on in the raid group undergoing replacement.
	if ok, err := bdr.IsExistingReplacmentInProgress(oldRG); ok {
		return false, fmt.Sprintf("cannot replace blockdevice as a "+
			"background replacement may be in progress in the raid group: %s", err.Error())
	}

	// The incoming BD should be a valid entry if
	// 1. The BD does not have a BDC.
	// 2. The BD has a BDC with the current CSPC label and there is no successor of this BD
	//    present in the CSPC.
	if !bdr.AreNewBDsValid(newRG, oldRG, bdr.OldCSPC) {
		return false, "the new blockdevice intended to use for replacement in invalid"
	}

	return true, ""
}

func (bdr *BlockDeviceReplacement) CreateBDCsForNewBDs() (bool, error) {
	newBDs := GetNewBDFromCSPC(bdr.NewCSPC, bdr.OldCSPC)

	for newBD, OldBD := range newBDs {
		err := createBDC(newBD, OldBD, bdr.OldCSPC)
		if err != nil {
			return false, errors.Errorf("could not create bdc for bd %s : %s", newBD, err.Error())
		}
	}
	return true, nil
}

// IsMoreThanOneDiskReplaced returns true if more than one disk is replaced in the same raid group.
func IsMoreThanOneDiskReplaced(newRG apis.RaidGroup, oldRG apis.RaidGroup) bool {
	count := GetNumberOfDiskReplaced(newRG, oldRG)

	if count == 2 {
		return true
	}
	return false
}

// IsNewBDPresentOnCurrentCSPC returns true if the new/incoming BD that will be used for replacement
// is already present in CSPC.
func (bdr *BlockDeviceReplacement) IsNewBDPresentOnCurrentCSPC(newRG apis.RaidGroup, oldRG apis.RaidGroup) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for _, pool := range bdr.OldCSPC.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				if _, ok := newBDs[bd.BlockDeviceName]; ok {
					return true
				}
			}
		}
	}
	return false
}

// IsExistingReplacmentInProgress returns true if a block device in raid group is under active replacement.
func (bdr *BlockDeviceReplacement) IsExistingReplacmentInProgress(oldRG apis.RaidGroup) (bool, error) {
	for _, v := range oldRG.BlockDevices {
		err, bdcObject := bdr.GetBDCOfBD(v.BlockDeviceName)
		if err != nil {
			return true, errors.Errorf("failed to query for any existing replacement in the raid group : %s", err.Error())
		}
		if bdcObject.HasAnnotationKey(apis.PredecessorBDKey) {
			return true, errors.Errorf("replacement is still in progress for bd %s", v.BlockDeviceName)
		}
	}
	return false, nil
}

// AreNewBDsValid returns true if the new BDs are valid BDs for replacement.
func (bdr *BlockDeviceReplacement) AreNewBDsValid(newRG apis.RaidGroup, oldRG apis.RaidGroup, oldcspc *apis.CStorPoolCluster) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for new, _ := range newBDs {
		err, bdc := bdr.GetBDCOfBD(new)
		if err != nil {
			return false
		}
		if !bdr.IsBDValid(new, bdc, oldcspc) {
			return false
		}
	}
	return true
}

// IsBDValid returns true if the new BD is a valid BD for replacement.
func (bdr *BlockDeviceReplacement) IsBDValid(bd string, bdc *bdc.BlockDeviceClaim, oldcspc *apis.CStorPoolCluster) bool {
	if bdc != nil && !bdc.HasLabel(string(apis.CStorPoolClusterCPK), oldcspc.Name) {
		return false
	}
	err, predecessorMap := bdr.GetPredecessorBDIfAny(oldcspc)
	if err != nil {
		return false
	}
	if predecessorMap[bd] {
		return false
	}
	return true
}

// GetPredecessorBDIfAny returns a map of predecessor BDs if any in the current CSPC
// Note: Predecessor BDs in a CSPC are those BD for which a new BD has appeared in the CSPC and
//       replacement is still in progress
//
// For example,
// (b1,b2) is a group in cspc
// which has been changed to ( b3,b2 )  [Notice that b1 got replaced by b3],
// now b1 is not present in CSPC but the replacement is still in progress in background.
// In this case b1 is a predecessor BD.
func (bdr *BlockDeviceReplacement) GetPredecessorBDIfAny(cspcOld *apis.CStorPoolCluster) (error, map[string]bool) {
	predecessorBDMap := make(map[string]bool)
	for _, pool := range cspcOld.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				err, bdc := bdr.GetBDCOfBD(bd.BlockDeviceName)
				if err != nil {
					return err, nil
				}
				predecessorBDMap[bdc.Object.GetAnnotations()[apis.PredecessorBDKey]] = true
			}
		}
	}
	return nil, predecessorBDMap
}

// GetNewBDFromCSPC returns a map of new successor bd to old bd for replacement in the CSPC.
func GetNewBDFromCSPC(newCSPC, oldCSPC *apis.CStorPoolCluster) map[string]string {
	newBDs := make(map[string]string)
	oldBDMap := make(map[string]bool)
	for _, pool := range oldCSPC.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				oldBDMap[bd.BlockDeviceName] = true
			}
		}
	}

	for i, pool := range newCSPC.Spec.Pools {
		for j, rg := range pool.RaidGroups {
			for k, bd := range rg.BlockDevices {
				if !oldBDMap[bd.BlockDeviceName] {
					newBDs[bd.BlockDeviceName] = oldCSPC.Spec.Pools[i].RaidGroups[j].BlockDevices[k].BlockDeviceName
				}
			}
		}
	}
	return newBDs
}

// getBDCOfBD returns the BDC object for corresponding BD.
func (bdr *BlockDeviceReplacement) GetBDCOfBD(bdName string) (error, *bdc.BlockDeviceClaim) {
	bdcList, err := bdc.NewKubeClient().List(v1.ListOptions{})
	if err != nil {
		return errors.Errorf("failed to list bdc: %s", err.Error()), nil
	}
	list := bdc.ListBuilderFromAPIList(bdcList).WithFilter(bdc.HasBD(bdName)).List()

	// If there is not BDC for a BD -- this means it an acceptable situation for BD replacement
	// The incoming BD finally will have a BDC created, hence no error is returned.
	if list.Len() == 0 {
		return nil, nil
	}

	if list.Len() != 1 {
		return errors.Errorf("did not get exact one bdc for the bd %s", bdName), nil
	}
	return nil, bdc.BuilderForAPIObject(&list.ObjectList.Items[0]).BDC
}

func createBDC(newBD string, oldBD string, cspcOld *apis.CStorPoolCluster) error {
	bdObj, err := bd.NewKubeClient().Get(newBD, v1.GetOptions{})
	if err != nil {
		return err
	}
	err = ClaimBD(bdObj, oldBD, cspcOld)
	if err != nil {
		return err
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func ClaimBD(newBdObj *ndmapis.BlockDevice, oldBD string, cspcOld *apis.CStorPoolCluster) error {
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-cstor-" + string(newBdObj.UID)).
		WithNamespace(cspcOld.Namespace).
		WithLabels(map[string]string{string(apis.CStorPoolClusterCPK): cspcOld.Name, apis.PredecessorBDKey: oldBD}).
		WithBlockDeviceName(newBdObj.Name).
		WithHostName(newBdObj.Labels[string(apis.HostNameCPK)]).
		WithCapacity(volume.ByteCount(newBdObj.Spec.Capacity.Storage)).
		WithCSPCOwnerReference(cspcOld).
		WithFinalizer(cspcv1alpha1.CSPCFinalizer).
		Build()

	if err != nil {
		return errors.Wrapf(err, "failed to build block device claim for bd {%s}", newBdObj.Name)
	}
	_, err = bdc.NewKubeClient().WithNamespace(cspcOld.Namespace).Create(newBDCObj.Object)
	if k8serror.IsAlreadyExists(err) {
		klog.Infof("BDC for BD {%s} already created", newBdObj.Name)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to create block device claim for bd {%s}", newBdObj.Name)
	}
	return nil
}

// GetNewBDFromRaidGroups returns a map of new successor bd to old bd for replacement in a raid group
func GetNewBDFromRaidGroups(newRG apis.RaidGroup, oldRG apis.RaidGroup) map[string]string {
	newBlockDeviceMap := make(map[string]string)
	oldBlockDevicesMap := make(map[string]bool)
	for _, bdOld := range oldRG.BlockDevices {
		oldBlockDevicesMap[bdOld.BlockDeviceName] = true
	}
	for i, newRG := range newRG.BlockDevices {
		if !oldBlockDevicesMap[newRG.BlockDeviceName] {
			newBlockDeviceMap[newRG.BlockDeviceName] = oldRG.BlockDevices[i].BlockDeviceName
		}
	}
	return newBlockDeviceMap
}
