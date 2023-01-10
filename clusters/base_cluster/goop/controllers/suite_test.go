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
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

/*
Ginkgo gist:
- Context = Describe = When. These three functions are synonymous, each exists for spec narrative flow.
-
*/

func TestAPIs(t *testing.T) {
	// RegisterFailHandler(Fail) is the single line of glue code connecting Ginkgo to Gomega.
	// If we were to avoid dot-imports this would read as gomega.RegisterFailHandler(ginkgo.Fail) - what we're doing here is telling our matcher library (Gomega) which function to call (Ginkgo's Fail) in the event a failure is detected.
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

// TestCRD smoke tests the CRD:
// - can it be installed without errors
var _ = Describe("CRD installation", func() {
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
	})

	It("Should delete the CRD successfully", func() {
		opts := envtest.CRDInstallOptions{
			Paths:              crdPaths,
			ErrorIfPathMissing: true,
			MaxTime:            5 * time.Second,
			PollInterval:       200 * time.Millisecond,
			CleanUpAfterUse:    true,
		}

		err := envtest.UninstallCRDs(cfg, opts)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Goop creation", func() {
	opts := envtest.CRDInstallOptions{
		Paths:              crdPaths,
		ErrorIfPathMissing: true,
		MaxTime:            5 * time.Second,
		PollInterval:       200 * time.Millisecond,
		CleanUpAfterUse:    true,
	}

	ctx := context.Background()

	goopName := "test-goop-creation"
	goop := &goopv1alpha1.Goop{
		ObjectMeta: metav1.ObjectMeta{
			Name: goopName,
		},
		Spec: goopv1alpha1.GoopSpec{
			Foo:      "blah",
			JobCount: 2,
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      goopName,
			Namespace: goopName,
		},
	}

	BeforeEach(func() {
		defs, err := envtest.InstallCRDs(cfg, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(defs)).Should(Equal(1))
		Expect(defs[0].Spec.Names.Kind).Should(Equal("Goop"))

		err = k8sClient.Create(ctx, namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := envtest.UninstallCRDs(cfg, opts)
		Expect(err).NotTo(HaveOccurred())

		_ = k8sClient.Delete(ctx, goop)
	})

	It("When Goop is created", func() {
		By("Creating Goop CR")
		err := k8sClient.Create(ctx, goop)
		Expect(err).NotTo(HaveOccurred())

		// TODO: additionl creation tests
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
