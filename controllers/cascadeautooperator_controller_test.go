package controllers

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cascadev1alpha1 "github.com/randsw/cascadeAuto-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Check if deployment and configmap created in cluster after CRD apply
var _ = Describe("CascadeAutoOperator controller", func() {
	Context("CascadeAutoOperator controller test", func() {

		const CascadeAutoOperatorName = "test-cascadeautooperator"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CascadeAutoOperatorName,
				Namespace: CascadeAutoOperatorName,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: CascadeAutoOperatorName, Namespace: CascadeAutoOperatorName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("CASCADEAUTOOPERATOR_IMAGE", "ghcr.io/randsw/cascadeautooperator:0.0.1")
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)

			By("Removing the Image ENV VAR which stores the Operand image")
			_ = os.Unsetenv("CASCADEAUTOOPERATOR_IMAGE")
		})

		It("should successfully reconcile a custom resource for CascadeAutoOperator", func() {
			By("Creating the custom resource for the Kind CascadeAutoOperator")
			cascadeauto := &cascadev1alpha1.CascadeAutoOperator{}
			err := k8sClient.Get(ctx, typeNamespaceName, cascadeauto)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				cascadeauto := &cascadev1alpha1.CascadeAutoOperator{
					ObjectMeta: metav1.ObjectMeta{
						Name:      CascadeAutoOperatorName,
						Namespace: namespace.Name,
						Labels: map[string]string{
							"app": "cascadeauto",
						},
					},
					Spec: cascadev1alpha1.CascadeAutoOperatorSpec{
						Replicas: 1,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "cascadescenario-test",
										Image: "ghcr.io/randsw/cascadescenariocontroller_v2:0.2.5",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "config-volume",
												MountPath: "/tmp",
											},
										},
										Env: []corev1.EnvVar{
											{
												Name: "POD_NAMESPACE",
												ValueFrom: &corev1.EnvVarSource{
													FieldRef: &corev1.ObjectFieldSelector{
														FieldPath: "metadata.namespace",
													},
												},
											},
										},
									},
								},
								Volumes: []corev1.Volume{
									{
										Name: "config-volume",
										VolumeSource: corev1.VolumeSource{
											ConfigMap: &corev1.ConfigMapVolumeSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: CascadeAutoOperatorName + "-cm",
												},
											},
										},
									},
								},
							},
						},
						ScenarioConfig: cascadev1alpha1.CascadeScenario{
							CascadeModules: []cascadev1alpha1.CascadeModule{
								{
									ModuleName: "grayscale",
									Configuration: map[string]string{
										"foo":   "bar",
										"spamm": "eggs",
										"test1": "test2",
									},
									Template: corev1.PodTemplateSpec{
										Spec: corev1.PodSpec{
											Containers: []corev1.Container{
												{
													Name:  "grayscale",
													Image: "ghcr.io/randsw/grayscale:0.1.1",
												},
											},
											RestartPolicy: corev1.RestartPolicyOnFailure,
										},
									},
								},
							},
						},
					},
				}

				err = k8sClient.Create(ctx, cascadeauto)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &cascadev1alpha1.CascadeAutoOperator{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			cascadeAutoOperatorReconciler := &CascadeAutoOperatorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = cascadeAutoOperatorReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if Deployment was successfully created in the reconciliation")
			Eventually(func() error {
				found := &appsv1.Deployment{}
				DeploymentName := CascadeAutoOperatorName + "-deploy"
				typeNamespaceName := types.NamespacedName{Name: DeploymentName, Namespace: CascadeAutoOperatorName}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			cascadeAutoOperatorReconciler = &CascadeAutoOperatorReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = cascadeAutoOperatorReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if ConfigMap was successfully created in the reconciliation")
			Eventually(func() error {
				found := &corev1.ConfigMap{}
				typeNamespaceName := types.NamespacedName{Name: CascadeAutoOperatorName + "-cm", Namespace: CascadeAutoOperatorName}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute*2, time.Second).Should(Succeed())
		})
	})
})
