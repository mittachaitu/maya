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
	"encoding/json"
	"fmt"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	bd "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	bdc "github.com/openebs/maya/pkg/blockdeviceclaim/v1alpha1"
	cspcv1alpha1 "github.com/openebs/maya/pkg/cstor/poolcluster/v1alpha1"
	"github.com/openebs/maya/pkg/volume"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"net/http"
)

// validateCSPC validates CSPC spec for Create, Update and Delete operation of the object.
func (wh *webhook) validateCSPC(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}
	// validates only if requested operation is CREATE or UPDATE
	if req.Operation == v1beta1.Update {
		klog.V(5).Infof("Admission webhook update request for type %s", req.Kind.Kind)
		return wh.validateCSPCUpdateRequest(req)
	} else if req.Operation == v1beta1.Create {
		klog.V(5).Infof("Admission webhook create request for type %s", req.Kind.Kind)
		return wh.validateCSPCCreateRequest(req)
	}

	klog.V(2).Info("Admission wehbook for PVC not " +
		"configured for operations other than UPDATE and CREATE")
	return response
}

// validateCSPCCreateRequest validates CSPC create request
func (wh *webhook) validateCSPCCreateRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().SetAllowed().WithResultAsSuccess(http.StatusAccepted).AR
	var cspc apis.CStorPoolCluster
	err := json.Unmarshal(req.Object.Raw, &cspc)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	if ok, msg := cspcValidation(&cspc); !ok {
		err := errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}
	return response
}

func cspcValidation(cspc *apis.CStorPoolCluster) (bool, string) {
	if len(cspc.Spec.Pools) == 0 {
		return false, fmt.Sprintf("pools in cspc should have at least one item")
	}

	for _, pool := range cspc.Spec.Pools {
		pool := pool // pin it
		ok, msg := poolSpecValidation(&pool)
		if !ok {
			return false, fmt.Sprintf("invalid pool spec: %s", msg)
		}
	}
	return true, ""
}

func poolSpecValidation(pool *apis.PoolSpec) (bool, string) {
	if pool.NodeSelector == nil || len(pool.NodeSelector) == 0 {
		return false, "nodeselector should not be empty"
	}
	if len(pool.RaidGroups) == 0 {
		return false, "at least one raid group should be present on pool spec"
	}
	// TODO : Add validation for pool config
	// Pool config will require mutating webhooks also.
	for _, raidGroup := range pool.RaidGroups {
		raidGroup := raidGroup // pin it
		ok, msg := raidGroupValidation(&raidGroup, &pool.PoolConfig)
		if !ok {
			return false, msg
		}
	}

	return true, ""
}

func raidGroupValidation(raidGroup *apis.RaidGroup, pool *apis.PoolConfig) (bool, string) {
	if raidGroup.Type == "" && pool.DefaultRaidGroupType == "" {
		return false, fmt.Sprintf("any one type at raid group or default raid group type be specified ")
	}
	if _, ok := apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)]; !ok {
		return false, fmt.Sprintf("unsupported raid type '%s' specified", apis.PoolType(raidGroup.Type))
	}

	if len(raidGroup.BlockDevices) == 0 {
		return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
	}

	if raidGroup.Type != string(apis.PoolStriped) {
		if len(raidGroup.BlockDevices) != apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)] {
			return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
		}
	} else {
		if len(raidGroup.BlockDevices) < apis.SupportedPRaidType[apis.PoolType(raidGroup.Type)] {
			return false, fmt.Sprintf("number of block devices honouring raid type should be specified")
		}
	}

	for _, bd := range raidGroup.BlockDevices {
		bd := bd
		ok, msg := blockDeviceValidation(&bd)
		if !ok {
			return false, msg
		}
	}
	return true, ""
}

func blockDeviceValidation(bd *apis.CStorPoolClusterBlockDevice) (bool, string) {
	if bd.BlockDeviceName == "" {
		return false, fmt.Sprint("block device name cannot be empty")
	}
	return true, ""
}

// validateCSPCUpdateRequest validates CSPC update request
// Note : Validation aspects for CSPC create and update are the same.
func (wh *webhook) validateCSPCUpdateRequest(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := NewAdmissionResponse().SetAllowed().WithResultAsSuccess(http.StatusAccepted).AR
	var cspcNew apis.CStorPoolCluster
	err := json.Unmarshal(req.Object.Raw, &cspcNew)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusBadRequest).AR
		return response
	}
	if ok, msg := cspcValidation(&cspcNew); !ok {
		err := errors.Errorf("invalid cspc specification: %s", msg)
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusUnprocessableEntity).AR
		return response
	}

	cspcOld, err := cspcv1alpha1.NewKubeClient().Get(cspcNew.Name, v1.GetOptions{})
	if err != nil {
		err := errors.Errorf("could not fetch existing cspc for validation: %s", err.Error())
		response = BuildForAPIObject(response).UnSetAllowed().WithResultAsFailure(err, http.StatusInternalServerError).AR
		return response
	}
	ValidateForBDReplacementCase(&cspcNew, cspcOld)

	return response
}

// IsBDReplacementValid validates for BD replacement.
func IsBDReplacementValid(newRG apis.RaidGroup, oldRG apis.RaidGroup, cspcOld *apis.CStorPoolCluster, cspcNew *apis.CStorPoolCluster) bool {

	// Not more than 1 bd should be replaced in a raid group.
	if IsMoreThanOneDiskReplaced(newRG, oldRG) {
		return false
	}

	// The incoming BD for replacement should not pe present in the current CSPC.
	if IsNewBDPresentOnCSPC(newRG, oldRG, cspcOld) {
		return false
	}

	// No background replacement should be going on in the raid group undergoing replacement.
	if IsExistingReplacmentInProgress(oldRG) {
		return false
	}

	// The incoming BD should be a valid entry if
	// 1. The BD does not have a BDC.
	// 2. The BD has a BDC with the current CSPC label and there is no successor of this BD
	//    present in the CSPC.
	if !AreNewBDsValid(newRG, oldRG, cspcOld) {
		return false
	}

	newBDs := GetNewBDFromCSPC(cspcOld, cspcNew)

	for _, bd := range newBDs {
		err := createBDC(bd, cspcOld)
		if err != nil {
			return false
		}
	}

	return true
}

func createBDC(bdName string, cspcOld *apis.CStorPoolCluster) error {
	bdObj, err := bd.NewKubeClient().Get(bdName, v1.GetOptions{})
	if err != nil {
		return err
	}
	err = ClaimBD(bdObj, cspcOld)
	if err != nil {
		return err
	}
	return nil
}

// ClaimBD claims a given BlockDevice
func ClaimBD(bdObj *ndmapis.BlockDevice, cspcOld *apis.CStorPoolCluster) error {
	newBDCObj, err := bdc.NewBuilder().
		WithName("bdc-cstor-" + string(bdObj.UID)).
		WithNamespace("openebs").
		WithLabels(map[string]string{string(apis.CStorPoolClusterCPK): cspcOld.Name,"openebs.io/bd-predecessor":""}).
		WithBlockDeviceName(bdObj.Name).
		WithHostName(bdObj.Labels[string(apis.HostNameCPK)]).
		WithCapacity(volume.ByteCount(bdObj.Spec.Capacity.Storage)).
		WithCSPCOwnerReference(cspcOld).
		WithFinalizer(cspcv1alpha1.CSPCFinalizer).
		Build()

	if err != nil {
		return errors.Wrapf(err, "failed to build block device claim for bd {%s}", bdObj.Name)
	}
	_, err = bdc.NewKubeClient().WithNamespace("openebs").Create(newBDCObj.Object)
	if k8serror.IsAlreadyExists(err) {
		klog.Infof("BDC for BD {%s} already created", bdObj.Name)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to create block device claim for bd {%s}", bdObj.Name)
	}
	return nil
}

// IsMoreThanOneDiskReplaced returns true if more than one disk is replaced in the same raid group.
func IsMoreThanOneDiskReplaced(newRG apis.RaidGroup, oldRG apis.RaidGroup) bool {
	count := GetNumberOfDiskReplaced(newRG, oldRG)

	if count == 2 {
		return true
	}
	return false
}

// IsExistingReplacmentInProgress returns true if a block device in raid group is under active replacement.
func IsExistingReplacmentInProgress(oldRG apis.RaidGroup) bool {
	for _, v := range oldRG.BlockDevices {
		err, bdcObject := getBDCOfBD(v.BlockDeviceName)
		if err != nil {
			return true
		}
		if bdcObject.HasAnnotationKey("openebs.io/bd-predecessor") {
			return true
		}
	}
	return false
}

// IsNewBDPresentOnCSPC returns true if the new/incoming BD that will be used for replacement
// is already present in CSPC.
func IsNewBDPresentOnCSPC(newRG apis.RaidGroup, oldRG apis.RaidGroup, oldcspc *apis.CStorPoolCluster) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for _, pool := range oldcspc.Spec.Pools {
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

// getBDCOfBD returns the BDC object for corresponding BD.
func getBDCOfBD(bdName string) (error, *bdc.BlockDeviceClaim) {
	bdcList, err := bdc.NewKubeClient().List(v1.ListOptions{})
	if err != nil {
		return err, nil
	}
	list := bdc.ListBuilderFromAPIList(bdcList).WithFilter(bdc.HasBD(bdName)).List()

	if list.Len() == 0 {
		return nil, nil
	}

	if list.Len() != 1 {
		return errors.New(""), nil
	}
	return nil, bdc.BuilderForAPIObject(&list.ObjectList.Items[0]).BDC
}

// IsBDValid returns true if the new BD is a valid BD for replacement.
func IsBDValid(bd string, bdc *bdc.BlockDeviceClaim, oldcspc *apis.CStorPoolCluster) bool {
	if bdc != nil && !bdc.HasLabel(string(apis.CStorPoolClusterCPK), oldcspc.Name) {
		return false
	}
	err, predecessorMap := GetPredecessorBDIfAny(oldcspc)
	if err != nil {
		return false
	}
	if predecessorMap[bd] {
		return false
	}
	return true
}

func GetPredecessorBDIfAny(cspcOld *apis.CStorPoolCluster) (error, map[string]bool) {
	predecessorBDMap := make(map[string]bool)
	for _, pool := range cspcOld.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				err, bdc := getBDCOfBD(bd.BlockDeviceName)
				if err != nil {
					return err, nil
				}
				predecessorBDMap[bdc.Object.GetAnnotations()["openebs.io/bd-predecessor"]] = true
			}
		}
	}
	return nil, predecessorBDMap
}

// AreNewBDsValid returns true if the new BDs are valid BDs for replacement.
func AreNewBDsValid(newRG apis.RaidGroup, oldRG apis.RaidGroup, oldcspc *apis.CStorPoolCluster) bool {
	newBDs := GetNewBDFromRaidGroups(newRG, oldRG)
	for new, _ := range newBDs {
		err, bdc := getBDCOfBD(new)
		if err != nil {
			return false
		}
		if !IsBDValid(new, bdc, oldcspc) {
			return false
		}
	}
	return true
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

// Get the new replacement BDs
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

// Get the new replacement BDs
func GetNewBDFromCSPC(oldCSPC, newCSPC *apis.CStorPoolCluster) []string {
	var newBDs []string
	oldBDMap := make(map[string]bool)
	for _, pool := range oldCSPC.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				oldBDMap[bd.BlockDeviceName] = true
			}
		}
	}

	for _, pool := range newCSPC.Spec.Pools {
		for _, rg := range pool.RaidGroups {
			for _, bd := range rg.BlockDevices {
				if !oldBDMap[bd.BlockDeviceName] {
					newBDs = append(newBDs, bd.BlockDeviceName)
				}
			}
		}
	}
	return newBDs
}

// ValidateForBDReplacementCase validates the changes in CSPC for block device replacement.
func ValidateForBDReplacementCase(cspcNew *apis.CStorPoolCluster, cspcOld *apis.CStorPoolCluster) bool{
	for i, pool := range cspcNew.Spec.Pools {
		for j, rg := range pool.RaidGroups {
			count := GetNumberOfDiskReplaced(rg, cspcOld.Spec.Pools[i].RaidGroups[j])
			if count > 1 {
				if !IsBDReplacementValid(rg, cspcOld.Spec.Pools[i].RaidGroups[j], cspcOld, cspcNew) {
					return false
				}
			}
		}
	}
	return true
}
