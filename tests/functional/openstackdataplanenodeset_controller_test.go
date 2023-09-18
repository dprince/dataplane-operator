/*
Copyright 2023.

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
package functional

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dataplanev1 "github.com/openstack-k8s-operators/dataplane-operator/api/v1beta1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	. "github.com/openstack-k8s-operators/lib-common/modules/common/test/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Dataplane Role Test", func() {
	var dataplaneNodeSetName types.NamespacedName
	var dataplaneSecretName types.NamespacedName
	var dataplaneSSHSecretName types.NamespacedName
	defaultEdpmServiceList := []string{
		"edpm_frr_image",
		"edpm_iscsid_image",
		"edpm_logrotate_crond_image",
		"edpm_nova_compute_image",
		"edpm_nova_libvirt_image",
		"edpm_ovn_controller_agent_image",
		"edpm_ovn_metadata_agent_image",
		"edpm_ovn_bgp_agent_image",
	}

	BeforeEach(func() {
		dataplaneNodeSetName = types.NamespacedName{
			Name:      "edpm-compute-nodeset",
			Namespace: namespace,
		}
		dataplaneSecretName = types.NamespacedName{
			Namespace: namespace,
			Name:      "dataplanenodeset-edpm-compute-nodeset",
		}
		dataplaneSSHSecretName = types.NamespacedName{
			Namespace: namespace,
			Name:      "dataplane-ansible-ssh-private-key-secret",
		}
		err := os.Setenv("OPERATOR_SERVICES", "../../config/services")
		Expect(err).NotTo(HaveOccurred())
	})

	When("A Dataplane resorce is created with PreProvisioned nodes", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateDataplaneNodeSet(dataplaneNodeSetName, DefaultDataPlaneNoNodeSetSpec()))
		})
		It("should have the Spec fields initialized", func() {
			dataplaneNodeSetInstance := GetDataplaneNodeSet(dataplaneNodeSetName)
			Expect(dataplaneNodeSetInstance.Spec.DeployStrategy.Deploy).Should(BeFalse())
		})

		It("should have the Status fields initialized", func() {
			dataplaneNodeSetInstance := GetDataplaneNodeSet(dataplaneNodeSetName)
			Expect(dataplaneNodeSetInstance.Status.Deployed).Should(BeFalse())
		})

		It("should have input not ready and unknown Conditions initialized", func() {
			th.ExpectCondition(
				dataplaneNodeSetName,
				ConditionGetterFunc(DataplaneConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				dataplaneNodeSetName,
				ConditionGetterFunc(DataplaneConditionGetter),
				condition.InputReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				dataplaneNodeSetName,
				ConditionGetterFunc(DataplaneConditionGetter),
				dataplanev1.SetupReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				dataplaneNodeSetName,
				ConditionGetterFunc(DataplaneConditionGetter),
				condition.DeploymentReadyCondition,
				corev1.ConditionUnknown,
			)
		})

		It("Should not have created a Secret", func() {
			th.AssertSecretDoesNotExist(dataplaneSecretName)
		})
	})

	When("A ssh secret is created", func() {

		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateDataplaneNodeSet(dataplaneNodeSetName, DefaultDataPlaneNoNodeSetSpec()))
			CreateSSHSecret(dataplaneSSHSecretName)
		})
		It("Should have created a Secret", func() {
			secret := th.GetSecret(dataplaneSecretName)
			Expect(secret.Data["inventory"]).Should(
				ContainSubstring("edpm-compute-nodeset"))
		})
	})

	When("No default service image is provided", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateDataplaneNodeSet(dataplaneNodeSetName, DefaultDataPlaneNoNodeSetSpec()))
			CreateSSHSecret(dataplaneSSHSecretName)
		})
		It("Should have default service values provided", func() {
			secret := th.GetSecret(dataplaneSecretName)
			for _, svcImage := range defaultEdpmServiceList {
				Expect(secret.Data["inventory"]).Should(
					ContainSubstring(svcImage))
			}
		})
	})

	When("A user provides a custom service image", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateDataplaneNodeSet(dataplaneNodeSetName, CustomServiceImageSpec()))
			CreateSSHSecret(dataplaneSSHSecretName)
		})
		It("Should have the user defined image in the inventory", func() {
			secret := th.GetSecret(dataplaneSecretName)
			Expect(secret.Data["inventory"]).Should(
				ContainSubstring("blah.test-image"))
		})
	})
})
