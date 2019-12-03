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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

type MockedValidBlockDeviceReplacement struct{}
type MockedInValidBlockDeviceReplacement struct{}

func (mbdr *MockedValidBlockDeviceReplacement) IsBDReplacementValid(newRG apis.RaidGroup, oldRG apis.RaidGroup) (bool, string) {
	return true, ""
}

func (mbdr *MockedInValidBlockDeviceReplacement) IsBDReplacementValid(newRG apis.RaidGroup, oldRG apis.RaidGroup) (bool, string) {
	return false, ""
}

func TestBlockDeviceReplacement_ValidateForBDReplacementCase(t *testing.T) {
	type fields struct {
		OldCSPC             *apis.CStorPoolCluster
		NewCSPC             *apis.CStorPoolCluster
		ValidBDR            MockedValidBlockDeviceReplacement
		InValidBDR          MockedInValidBlockDeviceReplacement
		ReplacementValidity bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Case#1: Not a case of block device replacement",
			fields: fields{
				OldCSPC: &apis.CStorPoolCluster{},
				NewCSPC: &apis.CStorPoolCluster{},
			},
			want: true,
		},
		{
			name: "Case#2: Not a case of block device replacement",
			fields: fields{
				OldCSPC: &apis.CStorPoolCluster{},
				NewCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dummyKey": "dummyValue",
						},
					},
				},
			},
			want: true,
		},

		{
			name: "Case#3: Not a case of block device replacement",
			fields: fields{
				OldCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-1"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-1"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				NewCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-0"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-1"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				ReplacementValidity: true,
			},
			want: true,
		},

		{
			name: "Case#4: A case of block device replacement with valid replacement",
			fields: fields{
				OldCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-1"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-1"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				NewCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-0"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-3"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				ReplacementValidity: true,
			},
			want: true,
		},

		{
			name: "Case#4: A case of block device replacement with invalid replacement",
			fields: fields{
				OldCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-1"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-1"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				NewCSPC: &apis.CStorPoolCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cspc-mirror",
						Namespace: "openebs",
					},
					Spec: apis.CStorPoolClusterSpec{
						Pools: []apis.PoolSpec{
							{
								NodeSelector: map[string]string{"kubernetes.io/hostname": "node-0"},
								RaidGroups: []apis.RaidGroup{
									{
										BlockDevices: []apis.CStorPoolClusterBlockDevice{
											{BlockDeviceName: "bd-3"},
											{BlockDeviceName: "bd-2"},
										},
									},
								},
							},
						},
					},
				},
				ReplacementValidity: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bdrv *BlockDeviceReplacementValidation
			if tt.fields.ReplacementValidity {
				bdrv = &BlockDeviceReplacementValidation{BlockDeviceReplacement: &MockedValidBlockDeviceReplacement{}}
			} else {
				bdrv = &BlockDeviceReplacementValidation{BlockDeviceReplacement: &MockedInValidBlockDeviceReplacement{}}
			}

			got, _ := ValidateForBDReplacementCase(tt.fields.NewCSPC, tt.fields.OldCSPC, bdrv)
			if got != tt.want {
				t.Errorf("BlockDeviceReplacement.ValidateForBDReplacementCase() got = %v, want %v", got, tt.want)
			}
		})
	}
}
