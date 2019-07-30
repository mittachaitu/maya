/*
Copyright 2019 The OpenEBS Authors
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

package reconciliation

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	spc "github.com/openebs/maya/pkg/storagepoolclaim/v1alpha1"
	"github.com/openebs/maya/tests/artifacts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("STRIPED SPARSE SPC", func() {

	When("We apply sparse-striped-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			Spc = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build()
			spcObj = Spc.Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))

			Expect(ops.IsSPCFinalizerExistsOnSPC(spcObj.Name, spc.SPCFinalizer)).To(BeTrue())

			Expect(ops.IsSPCFinalizerExistsOnBDCs(metav1.ListOptions{
				LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcObj.Name,
			}, spc.SPCFinalizer)).To(BeTrue())
		})
	})

	When("Remove finalizer", func() {
		It("make sure finalizer comes back as part of reconcilation", func() {
			err := Spc.RemoveFinalizer(spc.SPCFinalizer)
			Expect(err).To(BeNil())
			Expect(ops.IsSPCFinalizerExistsOnSPC(spcObj.Name, spc.SPCFinalizer)).To(BeTrue())
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #5 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We update spc maxPool field to 1", func() {
		It("3 cstor pool should be present as down scaling is not supported ", func() {
			// Get the latest spc
			newSPC, err := ops.SPCClient.Get(spcObj.Name, metav1.GetOptions{})
			Expect(err).To(BeNil())

			// update the spc to set maxPool field to 1
			obj := spc.BuilderForAPIObject(newSPC).WithMaxPool(1)
			_, err = ops.SPCClient.Update(obj.Spc.Object)
			Expect(err).To(BeNil())

			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 1)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			_, err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
			bdcCount := ops.GetBDCCount(
				metav1.ListOptions{
					LabelSelector: string(apis.StoragePoolClaimCPK) + "=" + spcObj.Name},
				0, string(artifacts.OpenebsNamespace))
			Expect(bdcCount).To(BeZero())
			Expect(ops.IsSPCNotExists(spcObj.Name)).To(BeTrue())
		})
	})
})

var _ = Describe("MIRRORED SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-mirrored-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeMirroredCPV)).
				Build().Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			_, err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})

var _ = Describe("RAIDZ SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-raidz-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {

			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeRaidzCPV)).
				Build().Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			_, err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})

var _ = Describe("RAIDZ2 SPARSE SPC", func() {

	// Test Case #1 (sparse-striped-auto-spc). | TestType : Pool Creation
	When("We apply sparse-raidz2-auto spc yaml with maxPool count equal to 3 on a k8s cluster having at least 3 capable node", func() {
		It("pool resources count should be 3 with no error and healthy status", func() {
			spcObj = spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(3).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeRaidz2CPV)).
				Build().Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #2 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 1 pool out of 3 by deleting one of the csp", func() {
		It("a new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 1)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #3 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 2 pool out of 3 by deleting 2 csps", func() {
		It("2 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 2)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	//Test Case #4 : Dependent on above test case #1 . | TestType : Reconciliation
	When("We delete 3 pool out of 3 by deleting 3 csps", func() {
		It("3 new pool should come up again", func() {
			ops.DeleteCSP(spcObj.Name, 3)
			// We expect 3 cstorPool objects.
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 3)
			Expect(cspCount).To(Equal(3))
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			_, err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})

})

var _ = Describe("SPC POOL POD DELETION", func() {
	When("We apply sparse-striped-auto spc yaml with maxPool count equal to 1"+
		"on a k8s cluster having at least one capable node", func() {
		It("pool resources count should be 1 with no error and healthy status", func() {
			namespace := string(artifacts.OpenebsNamespace)
			builtSpcObj := spc.NewBuilder().
				WithGenerateName(spcName).
				WithDiskType(string(apis.TypeSparseCPV)).
				WithMaxPool(1).
				WithOverProvisioning(false).
				WithPoolType(string(apis.PoolTypeStripedCPV)).
				Build()
			spcObj = builtSpcObj.Object

			// Create a storage pool claim
			_, err := ops.SPCClient.Create(spcObj)
			Expect(err).To(BeNil())
			cspCount := ops.GetHealthyCSPCount(spcObj.Name, 1)
			Expect(cspCount).To(Equal(1))

			// get pool pod corresponding to above spc
			poolPodList, err := ops.PodClient.WithNamespace(namespace).
				List(metav1.ListOptions{
					LabelSelector: "openebs.io/storage-pool-claim=" + spcObj.Name,
				},
				)
			Expect(err).To(BeNil(), "failed to get list of pool pods")

			// verify pod status
			status := ops.IsPodRunningEventually(namespace, poolPodList.Items[0].Name)
			Expect(status).To(Equal(true), "while checking status of pool pod {%s}", poolPodList.Items[0].Name)

			// Exec into cstor-pool pod and get pool guid
			oldOutput, err := ops.PodClient.WithNamespace(namespace).
				Exec(
					poolPodList.Items[0].Name,
					&corev1.PodExecOptions{
						Command: []string{
							"/bin/bash",
							"-c",
							"zpool get guid -H | awk '{print $3}'",
						},
						Container: poolPodList.Items[0].Spec.Containers[0].Name,
						Stdin:     false,
						Stdout:    true,
						Stderr:    true,
					},
				)
			Expect(err).To(BeNil(), "while geting the zpool pool guid of pod {%s} before restart", poolPodList.Items[0].Name)

			newPoolPodObj, err := ops.RestartPodEventually(
				namespace,
				metav1.ListOptions{
					LabelSelector: "openebs.io/storage-pool-claim=" + spcObj.Name,
				},
			)
			Expect(err).To(BeNil(), "failed to restart cstor pool pod")

			// Exec into cstor-pool pod and get pool guid
			newOutput, err := ops.PodClient.WithNamespace(namespace).
				Exec(
					newPoolPodObj.Name,
					&corev1.PodExecOptions{
						Command: []string{
							"/bin/bash",
							"-c",
							"zpool get guid -H | awk '{print $3}'",
						},
						Container: newPoolPodObj.Spec.Containers[0].Name,
						Stdin:     false,
						Stdout:    true,
						Stderr:    true,
					},
				)
			Expect(err).To(BeNil(), "while geting the zpool pool guid of pod {%s} after restart", newPoolPodObj)

			//Check zpool pool guid before and after restarts
			Expect(oldOutput.Stdout).To(Equal(newOutput.Stdout), "pool is created after restarting the pool pod")
		})
	})

	When("Cleaning up spc", func() {
		It("should delete the spc", func() {
			_, err := ops.SPCClient.Delete(spcObj.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		})
	})
})
