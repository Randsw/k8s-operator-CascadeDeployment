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
	"encoding/json"
	"fmt"
	"reflect"

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	cascadev1alpha1 "github.com/randsw/cascadeAuto-operator/api/v1alpha1"
	"github.com/randsw/cascadeAuto-operator/monitoring"
)

// CascadeAutoOperatorReconciler reconciles a CascadeAutoOperator object
type CascadeAutoOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Finalizer for metrics scrape
const finalizer = "metrics.cascade.cascade.net/finalizer"

//+kubebuilder:rbac:groups=cascade.cascade.net,resources=cascadeautooperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cascade.cascade.net,resources=cascadeautooperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cascade.cascade.net,resources=cascadeautooperators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CascadeAutoOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *CascadeAutoOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("CascadeAutoOperator", req.NamespacedName)

	logger.Info("Reconciling CascadeAutoOperator", "request name", req.Name, "request namespace", req.Namespace)

	instance := &cascadev1alpha1.CascadeAutoOperator{}

	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get CascadeAutoOperator.")
		return ctrl.Result{}, err
	}

	// Validate the CR spec before any processing
	if errs := r.validateSpec(instance); len(errs) > 0 {
		for _, verr := range errs {
			logger.Error(verr, "CR spec validation failed")
		}
		instance.Status.Result = fmt.Sprintf("validation failed: %v", errs)
		if err := r.Status().Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update status after validation failure")
		}
		// Don't requeue - invalid spec won't fix itself without user intervention
		return ctrl.Result{}, nil
	}

	// Add finalizer for metrics
	if !controllerutil.ContainsFinalizer(instance, finalizer) {
		logger.Info("Adding Finalizer for CascadeAutoOperator")
		controllerutil.AddFinalizer(instance, finalizer)
		if err = r.Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to update custom resource to add finalizer")
			return ctrl.Result{}, err
		}
	}
	isApplicationMarkedToBeDeleted := instance.GetDeletionTimestamp() != nil
	if isApplicationMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(instance, finalizer) {
			r.finalizeApplication(ctx, instance)
			controllerutil.RemoveFinalizer(instance, finalizer)
			if err := r.Update(ctx, instance); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				// Requeue to retry finalizer removal
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, nil
	}

	// Check if the Deployment already exists, if not create a new one
	found := &apps.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name + "-deploy", Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		deployment, err := r.createDeployment(instance, ctx, &logger)
		if err != nil {
			logger.Error(err, "Failed to create Deployment spec")
			return ctrl.Result{}, err
		}
		logger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
		// Increment instance count only after successful creation
		monitoring.SafeIncrement()
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Deployment exists - check if update is needed
	desiredDeployment, err := r.createDeployment(instance, ctx, &logger)
	if err != nil {
		logger.Error(err, "Failed to create desired Deployment spec for comparison")
		return ctrl.Result{}, err
	}
	if !deploymentSpecEqual(&found.Spec, &desiredDeployment.Spec) {
		logger.Info("Updating existing Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		found.Spec = desiredDeployment.Spec
		err = r.Update(ctx, found)
		if err != nil {
			logger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	foundMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name + "-cm", Namespace: instance.Namespace}, foundMap)
	if err != nil && errors.IsNotFound(err) {
		cm, err := r.getCm(instance, &logger)
		if err != nil {
			logger.Error(err, "Failed to create ConfigMap spec")
			return ctrl.Result{}, err
		}
		logger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
		err = r.Create(ctx, cm)
		if err != nil {
			logger.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// ConfigMap exists - check if update is needed
	desiredCm, err := r.getCm(instance, &logger)
	if err != nil {
		logger.Error(err, "Failed to create desired ConfigMap spec for comparison")
		return ctrl.Result{}, err
	}
	if !reflect.DeepEqual(foundMap.Data, desiredCm.Data) || !reflect.DeepEqual(foundMap.Labels, desiredCm.Labels) {
		logger.Info("Updating existing ConfigMap", "ConfigMap.Namespace", foundMap.Namespace, "ConfigMap.Name", foundMap.Name)
		foundMap.Data = desiredCm.Data
		foundMap.Labels = desiredCm.Labels
		err = r.Update(ctx, foundMap)
		if err != nil {
			logger.Error(err, "Failed to update ConfigMap", "ConfigMap.Namespace", foundMap.Namespace, "ConfigMap.Name", foundMap.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	foundSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		svc, err := r.getService(instance, &logger)
		if err != nil {
			logger.Error(err, "Failed to create Service spec")
			return ctrl.Result{}, err
		}
		logger.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			logger.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	// Service exists - check if update is needed
	desiredSvc, err := r.getService(instance, &logger)
	if err != nil {
		logger.Error(err, "Failed to create desired Service spec for comparison")
		return ctrl.Result{}, err
	}
	if !serviceSpecEqual(&foundSvc.Spec, &desiredSvc.Spec) || !reflect.DeepEqual(foundSvc.Annotations, desiredSvc.Annotations) || !reflect.DeepEqual(foundSvc.Labels, desiredSvc.Labels) {
		logger.Info("Updating existing Service", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)
		foundSvc.Spec = desiredSvc.Spec
		foundSvc.Annotations = desiredSvc.Annotations
		foundSvc.Labels = desiredSvc.Labels
		err = r.Update(ctx, foundSvc)
		if err != nil {
			logger.Error(err, "Failed to update Service", "Service.Namespace", foundSvc.Namespace, "Service.Name", foundSvc.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	// Update status to reflect successful reconciliation
	instance.Status.Result = "reconciliation succeeded"
	instance.Status.Active = found.Status.ReadyReplicas
	if err := r.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "Failed to update status after successful reconciliation")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CascadeAutoOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cascadev1alpha1.CascadeAutoOperator{}).
		Owns(&apps.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// validateSpec performs comprehensive validation of the CascadeAutoOperator CR spec
// before any resource processing occurs. Returns a slice of validation errors;
// an empty slice means the spec is valid.
func (r *CascadeAutoOperatorReconciler) validateSpec(instance *cascadev1alpha1.CascadeAutoOperator) []error {
	var errs []error

	// Validate ScenarioConfig has at least one CascadeModule
	modules := instance.Spec.ScenarioConfig.CascadeModules
	if len(modules) == 0 {
		errs = append(errs, fmt.Errorf("scenarioconfig.cascademodules must contain at least one module"))
	}

	// Validate each CascadeModule has a non-empty ModuleName
	for i, mod := range modules {
		if mod.ModuleName == "" {
			errs = append(errs, fmt.Errorf("scenarioconfig.cascademodules[%d].modulename must not be empty", i))
		}
	}

	// Validate pod template has at least one container
	containers := instance.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		errs = append(errs, fmt.Errorf("template.spec.containers must contain at least one container"))
	}

	// Validate pod template has at least one volume for config mount
	volumes := instance.Spec.Template.Spec.Volumes
	if len(volumes) == 0 {
		errs = append(errs, fmt.Errorf("template.spec.volumes must contain at least one volume for config mount"))
	} else if volumes[0].ConfigMap == nil {
		errs = append(errs, fmt.Errorf("template.spec.volumes[0] must be a ConfigMap volume for config mount"))
	}

	return errs
}

func (r *CascadeAutoOperatorReconciler) createDeployment(instance *cascadev1alpha1.CascadeAutoOperator, ctx context.Context, logger *logr.Logger) (*apps.Deployment, error) {
	ls := labelsForCascadeAutoOperator(instance.Name, instance.Name)
	replicas := instance.Spec.Replicas

	var podSpec = instance.Spec.Template

	podSpec.Labels = ls

	if len(podSpec.Spec.Volumes) == 0 {
		return nil, fmt.Errorf("pod template must have at least one volume for config mount")
	}
	if podSpec.Spec.Volumes[0].ConfigMap == nil {
		return nil, fmt.Errorf("first volume must be a ConfigMap volume for config mount")
	}
	podSpec.Spec.Volumes[0].ConfigMap.Name = instance.Name + "-cm"
	// Use special service account for cascade scenarion controller. SA created by heml-chart
	podSpec.Spec.ServiceAccountName = "cascade-scenario"

	dep := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deploy",
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: podSpec, // PodSec
		}, // Spec
	} // Deployment

	// Set CascadeAutoOperator instance as the owner and controller
	err := ctrl.SetControllerReference(instance, dep, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set CascadeAutoOperator instance as the owner and controller")
		return nil, err
	}
	return dep, nil
}

// Create configmap for scenario controller
func (r *CascadeAutoOperatorReconciler) getCm(instance *cascadev1alpha1.CascadeAutoOperator, logger *logr.Logger) (*corev1.ConfigMap, error) {
	data, err := json.Marshal(instance.Spec.ScenarioConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scenario config: %w", err)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-cm",
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
		},
		Data: map[string]string{
			"configuration": string(data),
		},
	}

	err = ctrl.SetControllerReference(instance, cm, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set CascadeAutoOperator instance as the owner and for configMap")
		return nil, err
	}
	return cm, nil
}

// Create service for scenario controller
func (r *CascadeAutoOperatorReconciler) getService(instance *cascadev1alpha1.CascadeAutoOperator, logger *logr.Logger) (*corev1.Service, error) {
	if len(instance.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("pod template must have at least one container")
	}
	var source string
	for _, envar := range instance.Spec.Template.Spec.Containers[0].Env {
		if envar.Name == "SID" {
			source = envar.Value
		}
	}
	var port []corev1.ServicePort
	port = append(port, corev1.ServicePort{Name: "http", Protocol: "TCP", Port: 80, TargetPort: intstr.IntOrString{IntVal: 8080}})
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			Labels:    instance.Labels,
			Annotations: map[string]string{
				"source": source,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: port,
			Selector: map[string]string{
				"app": instance.Name,
			},
		},
	}

	err := ctrl.SetControllerReference(instance, svc, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set CascadeAutoOperator instance as the owner and controller for service")
		return nil, err
	}
	return svc, nil
}

func labelsForCascadeAutoOperator(name_app string, name_cr string) map[string]string {
	return map[string]string{"app": name_app, "cascadeautooperator_cr": name_cr}
}

// deploymentSpecEqual uses semantic equality to compare Deployment specs,
// ignoring API server defaults (e.g. Strategy, RevisionHistoryLimit).
func deploymentSpecEqual(a, b *apps.DeploymentSpec) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return apiequality.Semantic.DeepDerivative(b, a)
}

// serviceSpecEqual uses semantic equality to compare Service specs,
// ignoring API server defaults (e.g. ClusterIP, SessionAffinity).
func serviceSpecEqual(a, b *corev1.ServiceSpec) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return apiequality.Semantic.DeepDerivative(b, a)
}

func (reconciler *CascadeAutoOperatorReconciler) finalizeApplication(ctx context.Context, application *cascadev1alpha1.CascadeAutoOperator) {
	// SafeDecrement protects the metrics gauge from going negative
	monitoring.SafeDecrement()
}
