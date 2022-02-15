/*
Copyright 2021 Red Hat, Inc.

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
	"fmt"

	gofakeit "github.com/brianvoe/gofakeit/v6"
	"github.com/go-logr/logr"
	taskrunapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	// "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	// "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	// "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// "sigs.k8s.io/controller-runtime/pkg/source"
)

// PipelineWatcherReconciler reconciles a watcher for pipelineRun objects
type PipelineWatcherReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Log              logr.Logger
	PipelineRunCache map[string]taskrunapi.PipelineRun
}

// Notice: We don't use permission markers for this controller

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *PipelineWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("PipelineRunWatcher", req.NamespacedName)

	// Get the Application resource
	var pipelineRun taskrunapi.PipelineRun
	err := r.Get(ctx, req.NamespacedName, &pipelineRun)
	if err != nil {
		if k8sErrors.IsNotFound(err) { //deleted pipelinerun
			defer delete(r.PipelineRunCache, req.NamespacedName.Name)

			pvcSubPath := r.PipelineRunCache[req.NamespacedName.Name].Spec.Workspaces[0].SubPath
			log.Info(pvcSubPath)

			deadline := int64(5400)
			cleanPvcPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "clean-pvc-pod" + pvcSubPath,
					Namespace: req.NamespacedName.Namespace,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:         "Never",
					ActiveDeadlineSeconds: &deadline,
					Containers: []corev1.Container{
						{
							Name: "pvc-cleaner",
							Command: []string{
								"/bin/bash",
							},
							TTY: true,
							Args: []string{
								"-c",
								// "ls; rm -rf ./*", // remove content inside subpath folder...
								"echo " + pvcSubPath + "; rm -rf " + pvcSubPath,
							},
							Image: "registry.access.redhat.com/ubi8/ubi-minimal:8.4-210",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "source",
									MountPath: "/workspace/source",
									// SubPath: pvcSubPath,
								},
							},
							// WorkingDir: "/workspace/source/",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "source",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "app-studio-default-workspace",
								},
							},
						},
					},
				},
			}

			// fmt.Printf("clean pvc pod %v", cleanPvcPod)
			if err := r.Client.Create(context.TODO(), cleanPvcPod); err != nil {
				return reconcile.Result{}, err
			}

			return reconcile.Result{Requeue: false}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))

	r.PipelineRunCache[pipelineRun.Name] = pipelineRun

	//    metadata:
	// 	labels:
	// 	  app: go-lang-test
	//   spec:
	// 	containers:
	// 	  - name: go-lang-test
	// 		command:
	// 		- bash
	// 		args:
	// 		- -c
	// 		- >-
	// 		  /application/folder-content-reader
	// 		image: quay.io/aandriienko/folder-reader
	// 		volumeMounts:
	// 		  - name: source
	// 			mountPath: /workspace/source
	// 	volumes:
	// 	  - name: source
	// 		persistentVolumeClaim:
	// 		  claimName: app-studio-default-workspace

	// cleanUpTask := taskrunapi.TaskRun{
	// 	ObjectMeta: v1.ObjectMeta{
	// 		Name:      "clean-workspaces-pvc-" + req.NamespacedName.Name, // todo generate unique name
	// 		Namespace: req.NamespacedName.Namespace,
	// 	},
	// 	Spec: taskrunapi.TaskRunSpec{
	// 		ServiceAccountName: "controller-manager",
	// 		TaskRef: &taskrunapi.TaskRef{
	// 			Kind: taskrunapi.ClusterTaskKind,
	// 			Name: "cleanup-build-directories",
	// 		},
	// 		// Timeout: &pipelineRun.Spec.Timeout{ },
	// 		Workspaces: []taskrunapi.WorkspaceBinding{
	// 			{
	// 				Name: "source",
	// 				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
	// 					ClaimName: "app-studio-default-workspace",
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// if err := r.Create(ctx, &cleanUpTask); err != nil {
	// 	return reconcile.Result{}, err
	// }

	log.Info(fmt.Sprintf("Finished reconcile loop for %v", req.NamespacedName))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	gofakeit.New(0)

	// mapPipelineRunFunc := func(obj client.Object) []reconcile.Request {
	// 	// todo: use labels to identify our pipelineruns
	// 	return []reconcile.Request{
	// 		{NamespacedName: types.NamespacedName{
	// 			Name:      obj.GetName(),
	// 			Namespace: obj.GetNamespace(),
	// 		}},
	// 	}
	// }

	deleteEventsPredicate := predicate.Funcs{
		UpdateFunc: func(evt event.UpdateEvent) bool {
			return false
		},
		CreateFunc: func(evt event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(evt event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(evt event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&taskrunapi.PipelineRun{}, builder.WithPredicates(deleteEventsPredicate)).
		// Watches(
		// 	&source.Kind{Type: &tektonapi.PipelineRun{}},
		// 	handler.EnqueueRequestsFromMapFunc(mapPipelineRunFunc),
		// 	builder.WithPredicates(deleteEventsPredicate)).
		Complete(r)
}
