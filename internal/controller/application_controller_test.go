/*
Copyright 2024.

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

package controller

import (
	argoprojv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Application Controller", func() {
	Context("When reconciling a resource", func() {

		It("should successfully reconcile the resource", func() {
			// labelがついていないApplicationを作成
			app := &argoprojv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app",
					Namespace: "default",
				},
				Spec: argoprojv1alpha1.ApplicationSpec{
					Destination: argoprojv1alpha1.ApplicationDestination{
						Namespace: "test",
					},
				},
			}

			Expect(k8sClient.Create(ctx, app)).Should(Succeed())

			createdApp := &argoprojv1alpha1.Application{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-app", Namespace: "default"}, createdApp)
			}).Should(Succeed())

			// finalizerが設定されていないことを確認
			Expect(createdApp.Finalizers).Should(BeEmpty())
		})
	})
})
