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
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// =============================================================================
// Unit Tests — helper functions (no envtest needed)
// =============================================================================

var _ = Describe("labelsForCascadeAutoOperator", func() {
	It("should return correct labels", func() {
		labels := labelsForCascadeAutoOperator("myapp", "mycr")
		Expect(labels).To(Equal(map[string]string{
			"app":                    "myapp",
			"cascadeautooperator_cr": "mycr",
		}))
	})

	It("should work with empty strings", func() {
		labels := labelsForCascadeAutoOperator("", "")
		Expect(labels).To(Equal(map[string]string{
			"app":                    "",
			"cascadeautooperator_cr": "",
		}))
	})
})

var _ = Describe("deploymentSpecEqual", func() {
	It("should return true for identical specs", func() {
		replicas := int32(3)
		spec := &appsv1.DeploymentSpec{Replicas: &replicas}
		Expect(deploymentSpecEqual(spec, spec)).To(BeTrue())
	})

	It("should return true when existing has API-server defaults absent from desired", func() {
		replicas := int32(3)
		desired := &appsv1.DeploymentSpec{Replicas: &replicas}
		existing := desired.DeepCopy()
		maxSurge := intstr.FromString("25%")
		maxUnavailable := intstr.FromString("25%")
		existing.Strategy = appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxSurge:       &maxSurge,
				MaxUnavailable: &maxUnavailable,
			},
		}
		var revHistory int32 = 10
		existing.RevisionHistoryLimit = &revHistory
		Expect(deploymentSpecEqual(existing, desired)).To(BeTrue(),
			"API-server defaults present in existing but absent from desired must be ignored")
	})

	It("should return false when replicas differ", func() {
		r1 := int32(3)
		r2 := int32(5)
		existing := &appsv1.DeploymentSpec{Replicas: &r1}
		desired := &appsv1.DeploymentSpec{Replicas: &r2}
		Expect(deploymentSpecEqual(existing, desired)).To(BeFalse())
	})

	It("should return false when template labels differ", func() {
		r := int32(1)
		existing := &appsv1.DeploymentSpec{
			Replicas: &r,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "old"},
				},
			},
		}
		desired := &appsv1.DeploymentSpec{
			Replicas: &r,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "new"},
				},
			},
		}
		Expect(deploymentSpecEqual(existing, desired)).To(BeFalse())
	})

	It("should return true for two nil specs", func() {
		Expect(deploymentSpecEqual(nil, nil)).To(BeTrue())
	})

	It("should return false when one spec is nil and the other is not", func() {
		replicas := int32(1)
		spec := &appsv1.DeploymentSpec{Replicas: &replicas}
		Expect(deploymentSpecEqual(nil, spec)).To(BeFalse())
		Expect(deploymentSpecEqual(spec, nil)).To(BeFalse())
	})
})

var _ = Describe("serviceSpecEqual", func() {
	It("should return true for identical specs", func() {
		spec := &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
			},
		}
		Expect(serviceSpecEqual(spec, spec)).To(BeTrue())
	})

	It("should return true when existing has ClusterIP and SessionAffinity absent from desired", func() {
		desired := &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
			},
		}
		existing := desired.DeepCopy()
		existing.ClusterIP = "10.0.0.1"
		existing.SessionAffinity = corev1.ServiceAffinityNone
		Expect(serviceSpecEqual(existing, desired)).To(BeTrue())
	})

	It("should return false when ports differ", func() {
		existing := &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
			},
		}
		desired := &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "https", Port: 443, Protocol: corev1.ProtocolTCP},
			},
		}
		Expect(serviceSpecEqual(existing, desired)).To(BeFalse())
	})

	It("should return false when selector differs", func() {
		existing := &corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{{Name: "http", Port: 80}},
			Selector: map[string]string{"app": "old"},
		}
		desired := &corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{{Name: "http", Port: 80}},
			Selector: map[string]string{"app": "new"},
		}
		Expect(serviceSpecEqual(existing, desired)).To(BeFalse())
	})

	It("should return true for two nil specs", func() {
		Expect(serviceSpecEqual(nil, nil)).To(BeTrue())
	})

	It("should return false when one spec is nil and the other is not", func() {
		spec := &corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "http", Port: 80}},
		}
		Expect(serviceSpecEqual(nil, spec)).To(BeFalse())
		Expect(serviceSpecEqual(spec, nil)).To(BeFalse())
	})
})

// =============================================================================
// Unit Tests — validateSpec (no envtest needed)
// =============================================================================

var _ = Describe("validateSpec", func() {
	var reconciler *CascadeAutoOperatorReconciler

	BeforeEach(func() {
		reconciler = &CascadeAutoOperatorReconciler{}
	})

	validModule := func() cascadev1alpha1.CascadeModule {
		return cascadev1alpha1.CascadeModule{
			ModuleName:    "grayscale",
			Configuration: map[string]string{"foo": "bar"},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{{Name: "module-ctr", Image: "test:latest"}},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		}
	}

	buildCR := func(modules []cascadev1alpha1.CascadeModule, containers []corev1.Container, volumes []corev1.Volume) *cascadev1alpha1.CascadeAutoOperator {
		return &cascadev1alpha1.CascadeAutoOperator{
			Spec: cascadev1alpha1.CascadeAutoOperatorSpec{
				ScenarioConfig: cascadev1alpha1.CascadeScenario{
					CascadeModules: modules,
				},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: containers,
						Volumes:    volumes,
					},
				},
			},
		}
	}

	validContainer := corev1.Container{Name: "main", Image: "app:latest"}

	validVolume := corev1.Volume{
		Name: "config-volume",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm"},
			},
		},
	}

	It("should pass validation for a completely valid spec", func() {
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{validModule()},
			[]corev1.Container{validContainer},
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(BeEmpty())
	})

	It("should pass validation for multiple valid modules", func() {
		m1 := validModule()
		m2 := validModule()
		m2.ModuleName = "invert"
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{m1, m2},
			[]corev1.Container{validContainer},
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(BeEmpty())
	})

	It("should fail when CascadeModules list is empty", func() {
		cr := buildCR(
			nil,
			[]corev1.Container{validContainer},
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("must contain at least one module"))
	})

	It("should fail when a module has an empty ModuleName", func() {
		bad := validModule()
		bad.ModuleName = ""
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{bad},
			[]corev1.Container{validContainer},
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("modulename must not be empty"))
	})

	It("should flag every module with an empty ModuleName", func() {
		bad1 := validModule()
		bad1.ModuleName = ""
		bad2 := validModule()
		bad2.ModuleName = ""
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{bad1, bad2},
			[]corev1.Container{validContainer},
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(2))
	})

	It("should fail when template containers list is empty", func() {
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{validModule()},
			nil,
			[]corev1.Volume{validVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("must contain at least one container"))
	})

	It("should fail when template volumes list is empty", func() {
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{validModule()},
			[]corev1.Container{validContainer},
			nil,
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("must contain at least one volume"))
	})

	It("should fail when the first volume is not a ConfigMap", func() {
		nonCMVolume := corev1.Volume{
			Name: "empty-dir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		cr := buildCR(
			[]cascadev1alpha1.CascadeModule{validModule()},
			[]corev1.Container{validContainer},
			[]corev1.Volume{nonCMVolume},
		)
		errs := reconciler.validateSpec(cr)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Error()).To(ContainSubstring("must be a ConfigMap volume"))
	})

	It("should collect all validation errors when multiple rules are violated", func() {
		cr := buildCR(
			nil, // no modules
			nil, // no containers
			nil, // no volumes
		)
		errs := reconciler.validateSpec(cr)
		Expect(len(errs)).To(BeNumerically(">=", 3),
			"expected at least 3 errors: missing modules, containers, and volumes")
	})
})

// =============================================================================
// Integration Tests — full reconciliation (envtest)
// =============================================================================

var _ = Describe("CascadeAutoOperator controller reconciliation", func() {
	var (
		ctx        context.Context
		crKey      types.NamespacedName
		namespace  *corev1.Namespace
		reconciler *CascadeAutoOperatorReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Use GenerateName so every test gets a unique namespace,
		// avoiding "object is being deleted" collisions.
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-cao-",
			},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

		crKey = types.NamespacedName{Name: namespace.Name, Namespace: namespace.Name}

		Expect(os.Setenv("CASCADEAUTOOPERATOR_IMAGE", "ghcr.io/randsw/cascadeautooperator")).To(Succeed())

		reconciler = &CascadeAutoOperatorReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	AfterEach(func() {
		_ = k8sClient.Delete(ctx, namespace)
		_ = os.Unsetenv("CASCADEAUTOOPERATOR_IMAGE")
	})

	// buildValidCR returns a fully populated, spec-valid CR used by multiple tests.
	// Must be called inside an It block so crKey is already initialised.
	buildValidCR := func() *cascadev1alpha1.CascadeAutoOperator {
		crName := crKey.Name
		return &cascadev1alpha1.CascadeAutoOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crName,
				Namespace: crKey.Namespace,
				Labels:    map[string]string{"app": "cascadeauto"},
			},
			Spec: cascadev1alpha1.CascadeAutoOperatorSpec{
				Replicas: 1,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app":                    crName,
							"cascadeautooperator_cr": crName,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "cascadescenario-test",
								Image: "ghcr.io/randsw/cascadescenariocontroller-auto",
								VolumeMounts: []corev1.VolumeMount{
									{Name: "config-volume", MountPath: "/tmp"},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "config-volume",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: crName + "-cm",
										},
									},
								},
							},
						},
						ServiceAccountName: "cascade-scenario",
					},
				},
				ScenarioConfig: cascadev1alpha1.CascadeScenario{
					CascadeModules: []cascadev1alpha1.CascadeModule{
						{
							ModuleName:    "grayscale",
							Configuration: map[string]string{"foo": "bar", "spamm": "eggs"},
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{Name: "grayscale", Image: "ghcr.io/randsw/grayscale:0.1.1"},
									},
									RestartPolicy: corev1.RestartPolicyOnFailure,
								},
							},
						},
					},
				},
			},
		}
	}

	// createAndReconcileOnce creates the CR and runs a single reconcile.
	createAndReconcileOnce := func(cr *cascadev1alpha1.CascadeAutoOperator) {
		if cr != nil {
			Expect(k8sClient.Create(ctx, cr)).To(Succeed())
		}
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
	}

	// ——————————————————————————————
	// Happy-path: resource creation
	// ——————————————————————————————

	It("should create Deployment, ConfigMap and Service in sequence", func() {
		createAndReconcileOnce(buildValidCR())

		By("creating the Deployment on first reconcile")
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-deploy", Namespace: crKey.Namespace},
				&appsv1.Deployment{},
			)
		}, time.Minute, time.Second).Should(Succeed())

		By("creating the ConfigMap on second reconcile")
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace},
				&corev1.ConfigMap{},
			)
		}, time.Minute, time.Second).Should(Succeed())

		By("creating the Service on third reconcile")
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name, Namespace: crKey.Namespace},
				&corev1.Service{},
			)
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("should set owner references on all created resources", func() {
		createAndReconcileOnce(buildValidCR())
		// Second reconcile
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		// Third reconcile
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		By("verifying Deployment owner reference")
		dep := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-deploy", Namespace: crKey.Namespace}, dep,
			)
		}, time.Minute, time.Second).Should(Succeed())
		Expect(dep.OwnerReferences).To(HaveLen(1))
		Expect(dep.OwnerReferences[0].Name).To(Equal(crKey.Name))

		By("verifying ConfigMap owner reference")
		cm := &corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace}, cm,
			)
		}, time.Minute, time.Second).Should(Succeed())
		Expect(cm.OwnerReferences).To(HaveLen(1))

		By("verifying Service owner reference")
		svc := &corev1.Service{}
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name, Namespace: crKey.Namespace}, svc,
			)
		}, time.Minute, time.Second).Should(Succeed())
		Expect(svc.OwnerReferences).To(HaveLen(1))
	})

	It("should populate ConfigMap data with the serialised scenario config", func() {
		createAndReconcileOnce(buildValidCR())
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		cm := &corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace}, cm,
			)
		}, time.Minute, time.Second).Should(Succeed())

		Expect(cm.Data).To(HaveKey("configuration"))
		Expect(cm.Data["configuration"]).To(ContainSubstring("grayscale"))
		Expect(cm.Data["configuration"]).To(ContainSubstring("\"foo\":\"bar\""))
	})

	It("should set the final status to 'reconciliation succeeded' after all resources exist", func() {
		createAndReconcileOnce(buildValidCR())
		// Reconcile until no more resources to create (4th call: all exist, no updates needed)
		for i := 0; i < 3; i++ {
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
			Expect(err).NotTo(HaveOccurred())
		}

		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Eventually(func() string {
			_ = k8sClient.Get(ctx, crKey, cr)
			return cr.Status.Result
		}, time.Minute, time.Second).Should(Equal("reconciliation succeeded"))
	})

	// ——————————————————————————————
	// Update tests
	// ——————————————————————————————

	It("should update the Deployment when replicas change", func() {
		createAndReconcileOnce(buildValidCR())
		// 2nd reconcile → create ConfigMap
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		// 3rd reconcile → create Service
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		By("patching the CR with updated replicas")
		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Expect(k8sClient.Get(ctx, crKey, cr)).To(Succeed())
		cr.Spec.Replicas = 3
		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		By("reconciling and checking the Deployment was updated")
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		dep := &appsv1.Deployment{}
		Eventually(func() int32 {
			_ = k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-deploy", Namespace: crKey.Namespace}, dep,
			)
			return *dep.Spec.Replicas
		}, time.Minute, time.Second).Should(Equal(int32(3)))
	})

	It("should update the ConfigMap when the scenario configuration changes", func() {
		createAndReconcileOnce(buildValidCR())
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		By("adding a new module and configuration key to the CR")
		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Expect(k8sClient.Get(ctx, crKey, cr)).To(Succeed())
		cr.Spec.ScenarioConfig.CascadeModules = append(cr.Spec.ScenarioConfig.CascadeModules,
			cascadev1alpha1.CascadeModule{
				ModuleName:    "invert",
				Configuration: map[string]string{"alpha": "beta"},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers:    []corev1.Container{{Name: "invert", Image: "invert:latest"}},
						RestartPolicy: corev1.RestartPolicyOnFailure,
					},
				},
			},
		)
		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		cm := &corev1.ConfigMap{}
		Eventually(func() string {
			_ = k8sClient.Get(ctx,
				types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace}, cm,
			)
			return cm.Data["configuration"]
		}, time.Minute, time.Second).Should(And(
			ContainSubstring("grayscale"),
			ContainSubstring("invert"),
			ContainSubstring("\"alpha\":\"beta\""),
		))
	})

	It("should update the Service when the CR labels change", func() {
		createAndReconcileOnce(buildValidCR())
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		By("changing CR labels which get propagated to the Service")
		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Expect(k8sClient.Get(ctx, crKey, cr)).To(Succeed())
		cr.Labels["app"] = "cascadeauto-updated"
		Expect(k8sClient.Update(ctx, cr)).To(Succeed())

		// The reconciler processes resources sequentially (Deployment → ConfigMap → Service).
		// Changing CR labels also affects the ConfigMap, so the first reconcile updates
		// the ConfigMap and requeues. A second reconcile is needed to reach the Service.
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		svc := &corev1.Service{}
		Eventually(func() map[string]string {
			_ = k8sClient.Get(ctx, crKey, svc)
			return svc.Labels
		}, time.Minute, time.Second).Should(HaveKeyWithValue("app", "cascadeauto-updated"))
	})

	// ——————————————————————————————
	// Validation error path
	// ——————————————————————————————

	It("should fail reconciliation gracefully when spec is invalid", func() {
		// Use a CR that passes CRD structural validation but fails the controller's
		// validateSpec. An empty ModuleName passes CRD validation (it's just a string)
		// but is rejected by the controller.
		invalidCR := buildValidCR()
		invalidCR.Spec.ScenarioConfig.CascadeModules[0].ModuleName = ""

		Expect(k8sClient.Create(ctx, invalidCR)).To(Succeed())

		result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred(), "validation errors must not return a Go error — they are user errors")
		Expect(result.Requeue).To(BeFalse(), "invalid spec must not be requeued")

		By("setting the status to reflect validation failure")
		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Expect(k8sClient.Get(ctx, crKey, cr)).To(Succeed())
		Expect(cr.Status.Result).To(ContainSubstring("validation failed"))
	})

	// ——————————————————————————————
	// Idempotency
	// ——————————————————————————————

	It("should be idempotent: reconcile after all resources exist does nothing", func() {
		createAndReconcileOnce(buildValidCR())
		for i := 0; i < 3; i++ {
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
			Expect(err).NotTo(HaveOccurred())
		}

		// Capture state
		dep := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx,
			types.NamespacedName{Name: crKey.Name + "-deploy", Namespace: crKey.Namespace}, dep,
		)).To(Succeed())
		oldRV := dep.ResourceVersion

		cm := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx,
			types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace}, cm,
		)).To(Succeed())
		oldCMRV := cm.ResourceVersion

		// Extra reconcile must not change anything
		result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())

		Expect(k8sClient.Get(ctx,
			types.NamespacedName{Name: crKey.Name + "-deploy", Namespace: crKey.Namespace}, dep,
		)).To(Succeed())
		Expect(dep.ResourceVersion).To(Equal(oldRV))

		Expect(k8sClient.Get(ctx,
			types.NamespacedName{Name: crKey.Name + "-cm", Namespace: crKey.Namespace}, cm,
		)).To(Succeed())
		Expect(cm.ResourceVersion).To(Equal(oldCMRV))
	})

	// ——————————————————————————————
	// Deletion / finalizer
	// ——————————————————————————————

	It("should add a finalizer and remove it on deletion", func() {
		createAndReconcileOnce(buildValidCR())

		By("having the finalizer added")
		cr := &cascadev1alpha1.CascadeAutoOperator{}
		Expect(k8sClient.Get(ctx, crKey, cr)).To(Succeed())
		Expect(cr.Finalizers).To(ContainElement("metrics.cascade.cascade.net/finalizer"))

		By("processing finalizer on deletion")
		Expect(k8sClient.Delete(ctx, cr)).To(Succeed())
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: crKey})
		Expect(err).NotTo(HaveOccurred())

		By("removing the finalizer so the resource is fully cleaned up")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, crKey, cr)
			if errors.IsNotFound(err) {
				return true
			}
			return len(cr.Finalizers) == 0
		}, time.Minute, time.Second).Should(BeTrue())
	})

	// ——————————————————————————————
	// Edge case: CR not found
	// ——————————————————————————————

	It("should return without error when the CR has already been deleted", func() {
		result, err := reconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: crKey.Namespace},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())
	})
})
