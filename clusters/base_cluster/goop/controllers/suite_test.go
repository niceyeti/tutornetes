/*
Copyright 2022.

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

package controllers

import (
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	goopv1alpha1 "github.com/example/goop/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
// Also see the example: https://github.com/kubernetes-sigs/kubebuilder/blob/v3.7.0/testdata/project-v3-with-deploy-image/controllers/busybox_controller_test.go
// And primarily: https://onsi.github.io/ginkgo/#getting-started

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var crdPaths []string = []string{filepath.Join("..", "config", "crd", "bases")}

func TestAPIs(t *testing.T) {
	// RegisterFailHandler(Fail) is the single line of glue code connecting Ginkgo to Gomega.
	// If we were to avoid dot-imports this would read as gomega.RegisterFailHandler(ginkgo.Fail) - what we're doing here is telling our matcher library (Gomega) which function to call (Ginkgo's Fail) in the event a failure is detected.
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

// TestCRD smoke tests the CRD:
// - can it be installed without errors
var _ = Describe("CRD installation", func() {
	Context("CRD installation test", func() {
		It("Should create the CRD successfully", func() {
			opts := envtest.CRDInstallOptions{
				Paths:              crdPaths,
				ErrorIfPathMissing: true,
				MaxTime:            5 * time.Second,
				PollInterval:       200 * time.Millisecond,
				CleanUpAfterUse:    true,
			}

			defs, err := envtest.InstallCRDs(cfg, opts)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(defs)).Should(Equal(1))
			Expect(defs[0].Spec.Names.Kind).Should(Equal("Goop"))
			//err = envtest.WaitForCRDs()
		})
	})
})

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     crdPaths,
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = goopv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// This marker allows new schemas to be added here automatically when a new API is added to the project.
	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
