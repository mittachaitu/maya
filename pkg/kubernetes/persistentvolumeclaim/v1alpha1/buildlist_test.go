package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestListBuilderWithAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs))
			if mock.expectedPVCLen != len(b.list.items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.list.items))
			}
		})
	}
}

func TestListBuilderWithAPIObjects(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
		expectedErr    bool
	}{
		"PVC set 1": {[]string{}, 0, true},
		"PVC set 2": {[]string{"pvc1"}, 1, false},
		"PVC set 3": {[]string{"pvc1", "pvc2"}, 2, false},
		"PVC set 4": {[]string{"pvc1", "pvc2", "pvc3"}, 3, false},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b, errs := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).APIList()
			if mock.expectedErr && len(errs) == 0 {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectedErr && len(errs) > 0 {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if !mock.expectedErr && mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.availablePVCs, len(b.Items))
			}
		})
	}
}

func TestListBuilderToAPIList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs  []string
		expectedPVCLen int
	}{
		"PVC set 1":  {[]string{}, 0},
		"PVC set 2":  {[]string{"pvc1"}, 1},
		"PVC set 3":  {[]string{"pvc1", "pvc2"}, 2},
		"PVC set 4":  {[]string{"pvc1", "pvc2", "pvc3"}, 3},
		"PVC set 5":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4"}, 4},
		"PVC set 6":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5"}, 5},
		"PVC set 7":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6"}, 6},
		"PVC set 8":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7"}, 7},
		"PVC set 9":  {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8"}, 8},
		"PVC set 10": {[]string{"pvc1", "pvc2", "pvc3", "pvc4", "pvc5", "pvc6", "pvc7", "pvc8", "pvc9"}, 9},
	}
	for name, mock := range tests {

		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			b := ListBuilderForAPIObjects(fakeAPIPVCList(mock.availablePVCs)).List().ToAPIList()
			if mock.expectedPVCLen != len(b.Items) {
				t.Fatalf("Test %v failed: expected %v got %v", name, mock.expectedPVCLen, len(b.Items))
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	tests := map[string]struct {
		availablePVCs map[string]corev1.PersistentVolumeClaimPhase
		filteredPVCs  []string
		filters       PredicateList
	}{
		"PVC Set 1": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC5": corev1.ClaimBound, "PVC6": corev1.ClaimPending, "PVC7": corev1.ClaimLost},
			filteredPVCs:  []string{"PVC5"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 2": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC3": corev1.ClaimBound, "PVC4": corev1.ClaimBound},
			filteredPVCs:  []string{"PVC2", "PVC4"},
			filters:       PredicateList{IsBound()},
		},

		"PVC Set 3": {
			availablePVCs: map[string]corev1.PersistentVolumeClaimPhase{"PVC1": corev1.ClaimLost, "PVC2": corev1.ClaimPending, "PVC3": corev1.ClaimPending},
			filteredPVCs:  []string{},
			filters:       PredicateList{IsBound()},
		},
	}
	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			list := ListBuilderForAPIObjects(fakeAPIPVCListFromNameStatusMap(mock.availablePVCs)).WithFilter(mock.filters...).List()
			if len(list.items) != len(mock.filteredPVCs) {
				t.Fatalf("Test %v failed: expected %v got %v", name, len(mock.filteredPVCs), len(list.items))
			}
		})
	}
}
