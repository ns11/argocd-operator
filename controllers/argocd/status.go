// Copyright 2020 ArgoCD Operator Developers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argocd

import (
	"context"
	"reflect"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	argoprojv1a1 "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	"github.com/argoproj-labs/argocd-operator/controllers/argoutil"
)

// reconcileStatus will ensure that all of the Status properties are updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatus(cr *argoprojv1a1.ArgoCD) error {
	if err := r.reconcileStatusApplicationController(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusSSOConfig(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusDex(cr); err != nil {
		log.Error(err, "error reconciling dex status")
	}

	if err := r.reconcileStatusPhase(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusRedis(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusRepo(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusServer(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusHost(cr); err != nil {
		return err
	}

	if err := r.reconcileStatusNotifications(cr); err != nil {
		return err
	}

	return nil
}

// reconcileStatusApplicationController will ensure that the ApplicationController Status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusApplicationController(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	ss := newStatefulSetWithSuffix("application-controller", "application-controller", cr)
	if argoutil.IsObjectFound(r.Client, cr.Namespace, ss.Name, ss) {
		status = "Pending"

		if ss.Spec.Replicas != nil {
			if ss.Status.ReadyReplicas == *ss.Spec.Replicas {
				status = "Running"
			}
		}
	}

	if cr.Status.ApplicationController != status {
		cr.Status.ApplicationController = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusDex will ensure that the Dex status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusDex(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	deploy := newDeploymentWithSuffix("dex-server", "dex-server", cr)
	if argoutil.IsObjectFound(r.Client, cr.Namespace, deploy.Name, deploy) {
		status = "Pending"

		if deploy.Spec.Replicas != nil {
			if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
				status = "Running"
			}
		}
	}

	if cr.Status.Dex != status {
		cr.Status.Dex = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusSSOConfig will ensure that the SSOConfig status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusSSOConfig(cr *argoprojv1a1.ArgoCD) error {

	// set status to track ssoConfigLegalStatus so it is always up to date with latest ssoConfig situation
	status := ssoConfigLegalStatus

	if cr.Status.SSOConfig != status {
		cr.Status.SSOConfig = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusPhase will ensure that the Status Phase is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusPhase(cr *argoprojv1a1.ArgoCD) error {
	var phase string

	if cr.Status.ApplicationController == "Running" && cr.Status.Redis == "Running" && cr.Status.Repo == "Running" && cr.Status.Server == "Running" {
		phase = "Available"
	} else {
		phase = "Pending"
	}

	if cr.Status.Phase != phase {
		cr.Status.Phase = phase
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusRedis will ensure that the Redis status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusRedis(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	if !cr.Spec.HA.Enabled {
		deploy := newDeploymentWithSuffix("redis", "redis", cr)
		if argoutil.IsObjectFound(r.Client, cr.Namespace, deploy.Name, deploy) {
			status = "Pending"

			if deploy.Spec.Replicas != nil {
				if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
					status = "Running"
				}
			}
		}
	} else {
		ss := newStatefulSetWithSuffix("redis-ha-server", "redis-ha-server", cr)
		if argoutil.IsObjectFound(r.Client, cr.Namespace, ss.Name, ss) {
			status = "Pending"

			if ss.Status.ReadyReplicas == *ss.Spec.Replicas {
				status = "Running"
			}
		}
		// TODO: Add check for HA proxy deployment here as well?
	}

	if cr.Status.Redis != status {
		cr.Status.Redis = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusRepo will ensure that the Repo status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusRepo(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	deploy := newDeploymentWithSuffix("repo-server", "repo-server", cr)
	if argoutil.IsObjectFound(r.Client, cr.Namespace, deploy.Name, deploy) {
		status = "Pending"

		if deploy.Spec.Replicas != nil {
			if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
				status = "Running"
			}
		}
	}

	if cr.Status.Repo != status {
		cr.Status.Repo = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusServer will ensure that the Server status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusServer(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	deploy := newDeploymentWithSuffix("server", "server", cr)
	if argoutil.IsObjectFound(r.Client, cr.Namespace, deploy.Name, deploy) {
		status = "Pending"

		// TODO: Refactor these checks.
		if deploy.Spec.Replicas != nil {
			if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
				status = "Running"
			}
		}
	}

	if cr.Status.Server != status {
		cr.Status.Server = status
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusNotifications will ensure that the Notifications status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusNotifications(cr *argoprojv1a1.ArgoCD) error {
	status := "Unknown"

	deploy := newDeploymentWithSuffix("notifications-controller", "controller", cr)
	if argoutil.IsObjectFound(r.Client, cr.Namespace, deploy.Name, deploy) {
		status = "Pending"

		if deploy.Spec.Replicas != nil {
			if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
				status = "Running"
			}
		}
	}

	if cr.Status.NotificationsController != status {
		if !cr.Spec.Notifications.Enabled {
			cr.Status.NotificationsController = ""
		} else {
			cr.Status.NotificationsController = status
		}
		return r.Client.Status().Update(context.TODO(), cr)
	}
	return nil
}

// reconcileStatusHost will ensure that the host status is updated for the given ArgoCD.
func (r *ReconcileArgoCD) reconcileStatusHost(cr *argoprojv1a1.ArgoCD) error {
	cr.Status.Host = ""
	cr.Status.Phase = "Available"

	if (cr.Spec.Server.Route.Enabled || cr.Spec.Server.Ingress.Enabled) && IsRouteAPIAvailable() {
		route := newRouteWithSuffix("server", cr)

		// The Red Hat OpenShift ingress controller implementation is designed to watch ingress objects and create one or more routes
		// to fulfill the conditions specified.
		// But the names of such created route resources are randomly generated so it is better to identify the routes using Labels
		// instead of Name.
		// 1. If a user creates ingress on openshift, Ingress controller generates a route for the ingress with random name.
		// 2. If a user creates route on openshift, Ingress controller processes the route with provided name.
		routeList := &routev1.RouteList{}
		opts := &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/name": route.Name,
			}),
			Namespace: cr.Namespace,
		}

		if err := r.Client.List(context.TODO(), routeList, opts); err != nil {
			return err
		}

		if len(routeList.Items) == 0 {
			log.Info("argocd-server route requested but not found on cluster")
			return nil
		} else {
			route = &routeList.Items[0]
			// status.ingress not available
			if route.Status.Ingress == nil {
				cr.Status.Host = ""
				cr.Status.Phase = "Pending"
			} else {
				// conditions exist and type is RouteAdmitted
				if len(route.Status.Ingress[0].Conditions) > 0 && route.Status.Ingress[0].Conditions[0].Type == routev1.RouteAdmitted {
					if route.Status.Ingress[0].Conditions[0].Status == corev1.ConditionTrue {
						cr.Status.Host = route.Status.Ingress[0].Host
						cr.Status.Phase = "Available"
					} else {
						cr.Status.Host = ""
						cr.Status.Phase = "Pending"
					}
				} else {
					// no conditions are available
					if route.Status.Ingress[0].Host != "" {
						cr.Status.Host = route.Status.Ingress[0].Host
						cr.Status.Phase = "Available"
					} else {
						cr.Status.Host = "Unavailable"
						cr.Status.Phase = "Pending"
					}
				}
			}
		}
	} else if cr.Spec.Server.Ingress.Enabled {
		ingress := newIngressWithSuffix("server", cr)
		if !argoutil.IsObjectFound(r.Client, cr.Namespace, ingress.Name, ingress) {
			log.Info("argocd-server ingress requested but not found on cluster")
			return nil
		} else {
			if !reflect.DeepEqual(ingress.Status.LoadBalancer, corev1.LoadBalancerStatus{}) && len(ingress.Status.LoadBalancer.Ingress) > 0 {
				var s []string
				var hosts string
				for _, ingressElement := range ingress.Status.LoadBalancer.Ingress {
					if ingressElement.Hostname != "" {
						s = append(s, ingressElement.Hostname)
						continue
					} else if ingressElement.IP != "" {
						s = append(s, ingressElement.IP)
						continue
					}
				}
				hosts = strings.Join(s, ", ")
				cr.Status.Host = hosts
				cr.Status.Phase = "Available"
			}
		}
	}
	return r.Client.Status().Update(context.TODO(), cr)
}
