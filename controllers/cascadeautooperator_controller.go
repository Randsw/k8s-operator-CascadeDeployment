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

	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cascadev1alpha1 "github.com/randsw/cascadeAuto-operator/api/v1alpha1"
	"github.com/randsw/cascadeAuto-operator/monitoring"
)

// CascadeAutoOperatorReconciler reconciles a CascadeAutoOperator object
type CascadeAutoOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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
			err := r.Update(ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Check if the Deployment already exists, if not create a new one
	found := &apps.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name + "-deploy", Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Deployment
		deployment := r.createDeployment(instance, ctx)
		// Increment instance count
		monitoring.CascadeAutoCurrentInstanceCount.Inc()
		logger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	foundMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name + "-cm", Namespace: instance.Namespace}, foundMap)
	if err != nil && errors.IsNotFound(err) {
		cm := r.getCm(instance)
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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CascadeAutoOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cascadev1alpha1.CascadeAutoOperator{}).
		Owns(&apps.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func (r *CascadeAutoOperatorReconciler) createDeployment(instance *cascadev1alpha1.CascadeAutoOperator, ctx context.Context) *apps.Deployment {
	logger := log.FromContext(ctx)
	ls := labelsForCascadeAutoOperator(instance.Name, instance.Name)
	replicas := instance.Spec.Replicas

	// Using the context to log information
	logger.Info("Logging: Creating a new Deployment", "Replicas", replicas)
	message := "Logging: (Name: " + instance.Name + "-deploy" + ") \n"
	logger.Info(message)
	message = "Logging: (Namespace: " + instance.Namespace + ") \n"
	logger.Info(message)

	for key, value := range ls {
		message = "Logging: (Key: [" + key + "] Value: [" + value + "]) \n"
		logger.Info(message)
	}

	var podSpec = instance.Spec.Template

	podSpec.Labels = ls

	podSpec.Spec.Volumes[0].ConfigMap.Name = instance.Name + "-cm"
	podSpec.Spec.ServiceAccountName = "cascade"

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
	ctrl.SetControllerReference(instance, dep, r.Scheme)
	return dep
}

func (r *CascadeAutoOperatorReconciler) getCm(instance *cascadev1alpha1.CascadeAutoOperator) *corev1.ConfigMap {
	data, _ := json.Marshal(instance.Spec.ScenarioConfig)
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

	ctrl.SetControllerReference(instance, cm, r.Scheme)
	return cm
}

func labelsForCascadeAutoOperator(name_app string, name_cr string) map[string]string {
	return map[string]string{"app": name_app, "cascadeautooperator_cr": name_cr}
}

func (reconciler *CascadeAutoOperatorReconciler) finalizeApplication(ctx context.Context, application *cascadev1alpha1.CascadeAutoOperator) {
	monitoring.CascadeAutoCurrentInstanceCount.Dec()
}
