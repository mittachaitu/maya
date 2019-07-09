/*
Copyright 2018 The OpenEBS Authors

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

package v1alpha1

import (
	"strconv"
	"testing"

	"github.com/golang/glog"
	ndmapis "github.com/openebs/maya/pkg/apis/openebs.io/ndm/v1alpha1"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha1"
	openebsFakeClientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned/fake"
	ndmFakeClientset "github.com/openebs/maya/pkg/client/generated/openebs.io/ndm/v1alpha1/clientset/internalclientset/fake"
	cstorpool "github.com/openebs/maya/pkg/cstor/pool/v1alpha1"
	sp "github.com/openebs/maya/pkg/sp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var blockDeviceK8sClient *blockdevice.KubernetesClient

func FakeDiskCreator(bd *blockdevice.KubernetesClient) {
	// Create some fake block device objects over nodes.
	// For example, create 14 disk (out of 14 disks, 2 disks are sparse disks)for each of 5 nodes.
	// That meant 14*5 i.e. 70 disk objects should be created

	// diskObjectList will hold the list of disk objects
	var diskObjectList [70]*ndmapis.BlockDevice

	sparseDiskCount := 2
	var key, diskLabel, deviceType string

	// nodeIdentifer will help in naming a node and attaching multiple disks to a single node.
	nodeIdentifer := 0
	for diskListIndex := 0; diskListIndex < 70; diskListIndex++ {
		diskIdentifier := strconv.Itoa(diskListIndex)
		if diskListIndex%14 == 0 {
			nodeIdentifer++
			sparseDiskCount = 0
		}
		if sparseDiskCount != 2 {
			deviceType = "sparse"
			sparseDiskCount++
		} else {
			deviceType = "disk"
		}
		key = "ndm.io/blockdevice-type"
		diskLabel = "blockdevice"
		diskObjectList[diskListIndex] = &ndmapis.BlockDevice{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "blockdevice" + diskIdentifier,
				Labels: map[string]string{
					"kubernetes.io/hostname": "gke-ashu-cstor-default-pool-a4065fd6-vxsh" + strconv.Itoa(nodeIdentifer),
					key:                      diskLabel,
				},
			},
			Spec: ndmapis.DeviceSpec{
				Details: ndmapis.DeviceDetails{
					DeviceType: deviceType,
				},
				Capacity: ndmapis.DeviceCapacity{
					Storage: 10737418240,
				},
				Partitioned: "NO",
				DevLinks: []ndmapis.DeviceDevLink{
					ndmapis.DeviceDevLink{
						Kind:  "by-ID",
						Links: []string{"/dev/sda"},
					},
				},
			},
			Status: ndmapis.DeviceStatus{
				State:      DiskStateActive,
				ClaimState: ndmapis.BlockDeviceUnclaimed,
			},
		}
		_, err := bd.Create(diskObjectList[diskListIndex])
		if err != nil {
			glog.Error(err)
		}
	}

}
func fakeDiskClient() {
	blockDeviceK8sClient = &blockdevice.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     ndmFakeClientset.NewSimpleClientset(),
		Namespace:     "fake-ns",
	}
}
func fakeAlgorithmConfig(spc *v1alpha1.StoragePoolClaim) *Config {
	var bdClient blockdevice.BlockDeviceInterface
	fakeDiskClient()
	FakeDiskCreator(blockDeviceK8sClient)
	if ProvisioningType(spc) == ProvisioningTypeManual {
		bdClient = &blockdevice.SpcObjectClient{
			KubernetesClient: blockDeviceK8sClient,
			Spc:              spc,
		}
	} else {
		bdClient = blockDeviceK8sClient
	}

	cspK8sClient := &cstorpool.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	spK8sClient := &sp.KubernetesClient{
		Kubeclientset: fake.NewSimpleClientset(),
		Clientset:     openebsFakeClientset.NewSimpleClientset(),
	}
	ac := &Config{
		Spc: spc,

		BlockDeviceClient: bdClient,
		CspClient:         cspK8sClient,
		SpClient:          spK8sClient,
		Namespace:         "fake-ns",
	}

	return ac
}

func TestProvisioningType(t *testing.T) {
	tests := map[string]struct {
		spc              *v1alpha1.StoragePoolClaim
		expectedPoolType string
	}{
		"autoSPC1": {
			spc: &v1alpha1.StoragePoolClaim{
				Spec: v1alpha1.StoragePoolClaimSpec{
					Type: "disk",
					PoolSpec: v1alpha1.CStorPoolAttr{
						PoolType: "striped",
					},
				},
			},
			expectedPoolType: ProvisioningTypeAuto,
		},
		"manualSPC2": {
			spc: &v1alpha1.StoragePoolClaim{
				Spec: v1alpha1.StoragePoolClaimSpec{
					Type: "disk",
					PoolSpec: v1alpha1.CStorPoolAttr{
						PoolType: "mirrored",
					},
					BlockDevices: v1alpha1.BlockDeviceAttr{
						BlockDeviceList: []string{"blockdevice1", "blockdevice2"},
					},
				},
			},
			expectedPoolType: ProvisioningTypeManual,
		},
		"autoSPC3": {
			spc: &v1alpha1.StoragePoolClaim{
				Spec: v1alpha1.StoragePoolClaimSpec{
					Type: "sparse",
					PoolSpec: v1alpha1.CStorPoolAttr{
						PoolType: "striped",
					},
				},
			},
			expectedPoolType: ProvisioningTypeAuto,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			mode := ProvisioningType(test.spc)
			if mode != test.expectedPoolType {
				t.Fatalf("Test %q failed expected mode: %s got %s", name, test.expectedPoolType, mode)
			}
		})
	}
}

//TODO: Uncomment or add UT for auto spc --> cspc
//func TestNodeBlockDeviceAlloter(t *testing.T) {
//	tests := map[string]struct {
//		// fakeCasPool holds the fake fakeCasPool object in test cases.
//		fakeCasPool *v1alpha1.StoragePoolClaim
//		// expectedDiskListLength holds the length of disk list
//		expectedDiskListLength int
//		expectedErr            bool
//	}{
//		// Test Case #1
//		"autoSPC1": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "striped",
//					},
//					MaxPools: func() *int {
//						count := 1
//						return &count
//					}(),
//				},
//			},
//			expectedDiskListLength: 1,
//			expectedErr:            false,
//		},
//		// Test Case #2
//		"autoSPC2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					MaxPools: func() *int {
//						count := 3
//						return &count
//					}(),
//				},
//			},
//			expectedDiskListLength: 6,
//			expectedErr:            false,
//		},
//		// Test Case #3
//		"autoSPC3": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "sparse",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "striped",
//					},
//					MaxPools: func() *int {
//						count := 4
//						return &count
//					}(),
//				},
//			},
//			expectedDiskListLength: 4,
//			expectedErr:            false,
//		},
//		// Test Case #4
//		"autoSPC4": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					MaxPools: func() *int {
//						count := 3
//						return &count
//					}(),
//				},
//			},
//			expectedDiskListLength: 18,
//			expectedErr:            false,
//		},
//		//Test Case #5
//		// blockdevice0, blockdevice1 are of sparse type
//		"manualSPC5": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "striped",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice0", "blockdevice1", "blockdevice2"},
//					},
//				},
//			},
//			expectedDiskListLength: 1,
//			expectedErr:            true,
//		},
//		// Test Case #6
//		"manualSPC6": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice3", "blockdevice4"},
//					},
//				},
//			},
//			expectedDiskListLength: 2,
//			expectedErr:            true,
//		},
//		// Test Case #7
//		"manualSPC7": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "sparse",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice71"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #8
//		"manualSPC8": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5"},
//					},
//				},
//			},
//			expectedDiskListLength: 4,
//			expectedErr:            true,
//		},
//		// Test Case #8
//		"manualSPC9": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3"},
//					},
//				},
//			},
//			expectedDiskListLength: 2,
//			expectedErr:            true,
//		},
//		// Test Case #10
//		"manualSPC10Raidz": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4"},
//					},
//				},
//			},
//			expectedDiskListLength: 3,
//			expectedErr:            true,
//		},
//		// Test Case #11
//		"manualSPC11Raidz": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice5", "blockdevice6"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #12
//		"manualSPC12Raidz": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4"},
//					},
//				},
//			},
//			expectedDiskListLength: 3,
//			expectedErr:            true,
//		},
//		// Test Case #13
//		"manualSPC13Raidz": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5"},
//					},
//				},
//			},
//			expectedDiskListLength: 3,
//			expectedErr:            true,
//		},
//		// Test Case #14
//		"manualSPC14Raidz": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice2"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #15
//		"manualSPC15Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #16
//		"manualSPC16Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #17
//		"manualSPC17Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #18
//		"manualSPC18Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6", "blockdevice7", "blockdevice8"},
//					},
//				},
//			},
//			expectedDiskListLength: 6,
//			expectedErr:            true,
//		},
//		// Test Case #19
//		"manualSPC19Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6", "blockdevice7", "blockdevice8", "blockdevice9", "blockdevice10", "blockdevice11", "blockdevice12", "blockdevice13"},
//					},
//				},
//			},
//			expectedDiskListLength: 12,
//			expectedErr:            true,
//		},
//		// Test Case #20
//		"manualSPC20Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice6", "blockdevice7"},
//					},
//				},
//			},
//			expectedDiskListLength: 6,
//			expectedErr:            true,
//		},
//		// Test Case #21
//		"manualSPC21Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "raidz2",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice1", "blockdevice2", "blockdevice3", "blockdevice4", "blockdevice5", "blockdevice27"},
//					},
//				},
//			},
//			expectedDiskListLength: 0,
//			expectedErr:            true,
//		},
//		// Test Case #22
//		"manualSPC22Raidz2": {
//			fakeCasPool: &v1alpha1.StoragePoolClaim{
//				Spec: v1alpha1.StoragePoolClaimSpec{
//					Type: "disk",
//					PoolSpec: v1alpha1.CStorPoolAttr{
//						PoolType: "mirrored",
//					},
//					BlockDevices: v1alpha1.BlockDeviceAttr{
//						BlockDeviceList: []string{"blockdevice0", "blockdevice1"},
//					},
//				},
//			},
//			expectedDiskListLength: 6,
//			expectedErr:            true,
//		},
//	}
//
//	for name, test := range tests {
//		name, test := name, test
//		t.Run(name, func(t *testing.T) {
//			//hostName := "gke-ashu-cstor-default-pool-a4065fd6-vxsh0"
//			ac := fakeAlgorithmConfig(test.fakeCasPool)
//			//blockdeviceList, err := ac.NodeBlockDeviceSelector()
//			selectedNodeBDs, err := ac.NodeBlockDeviceSelector()
//			if test.expectedErr && err == nil {
//				t.Fatalf("Test case failed expected error not to be nil")
//			}
//			if !test.expectedErr && err != nil {
//				t.Fatalf(
//					"Test case failed expected error to be nil but got error %v",
//					err,
//				)
//			}
//			if err == nil {
//				gotLen := 0
//				for _, val := range selectedNodeBDs {
//					val, _ := val.ToObjectList()
//					gotLen += len(val)
//				}
//				if gotLen != test.expectedDiskListLength {
//					t.Errorf("%q test case failed as the expected blockdevice list length is %d but got %d", name, test.expectedDiskListLength, gotLen)
//				}
//			}
//		})
//	}
//}
