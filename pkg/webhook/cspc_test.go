package webhook

import (
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
)

func TestIsMoreThanOneDiskReplaced(t *testing.T) {
	type args struct {
		newRG apis.RaidGroup
		oldRG apis.RaidGroup
	}
	tests := []struct {
		name string
		args args
		want bool
	}{

		{
			name: "No Replacement",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
			},
			want: false,
		},
		{
			name: "No Replacement",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-1"}},
				},
			},
			want: false,
		},
		{
			name: "1 Replacement",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-2"}},
				},
			},
			want: false,
		},
		{
			name: "2 Replacement",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}},
				},
			},
			want: true,
		},
		{
			name: "2 Replacement",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-3"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}, {BlockDeviceName: "bd-5"}},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMoreThanOneDiskReplaced(tt.args.newRG, tt.args.oldRG); got != tt.want {
				t.Errorf("IsMoreThanOneDiskReplaced() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNewBDFromRaidGroups(t *testing.T) {
	type args struct {
		newRG apis.RaidGroup
		oldRG apis.RaidGroup
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
		{
			name: "case#1",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}},
				},
			},
			want: map[string]string{
				"bd-1": "bd-3",
				"bd-2": "bd-4",
			},
		},

		{
			name: "case#2",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-6"}, {BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}},
				},
			},
			want: map[string]string{
				"bd-1": "bd-6",
			},
		},

		{
			name: "case#3",
			args: args{
				newRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-1"}, {BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-3"}, {BlockDeviceName: "bd-4"}},
				},
				oldRG: apis.RaidGroup{
					BlockDevices: []apis.CStorPoolClusterBlockDevice{{BlockDeviceName: "bd-6"}, {BlockDeviceName: "bd-2"}, {BlockDeviceName: "bd-7"}, {BlockDeviceName: "bd-4"}},
				},
			},
			want: map[string]string{
				"bd-1": "bd-6",
				"bd-3": "bd-7",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNewBDFromRaidGroups(tt.args.newRG, tt.args.oldRG); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNewBDFromRaidGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
