package ironic

import (
	"context"
    "reflect"
    "strconv"

	ironicv1alpha1 "github.com/metalkube/ironic-operator/pkg/apis/ironic/v1alpha1"
    helpers "github.com/metalkube/ironic-operator/pkg/helpers"

    appsv1 "k8s.io/api/apps/v1"
    batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ironic")

// Add creates a new Ironic Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIronic{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ironic-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Ironic
	err = c.Watch(&source.Kind{Type: &ironicv1alpha1.Ironic{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ironicv1alpha1.Ironic{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileIronic{}

// ReconcileIronic reconciles a Ironic object
type ReconcileIronic struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Ironic object and makes changes based on the state read
// and what is in the Ironic.Spec
func (r *ReconcileIronic) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Ironic")

	// Fetch the Ironic instance
	instance := &ironicv1alpha1.Ironic{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

    // Check if the configmap already exists, if not create a new one
    cm_found := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-bin", Namespace: instance.Namespace}, cm_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new configmap
        cm, _ := helpers.GetIronicBinConfigMap(instance.Namespace)
        reqLogger.Info("Creating a new ironic-bin configmap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
        err = r.client.Create(context.TODO(), cm)
        if err != nil {
            reqLogger.Error(err, "failed to create a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get ironic-bin ConfigMap")
        return reconcile.Result{}, err
    }

    cm_etc_found := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-etc", Namespace: instance.Namespace}, cm_etc_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new configmap
        cm_etc, _ := helpers.GetIronicEtcConfigMap(instance.Namespace, r.client)
        reqLogger.Info("Creating a new ironic-etc configmap", "ConfigMap.Namespace", cm_etc.Namespace, "ConfigMap.Name", cm_etc.Name)
        err = r.client.Create(context.TODO(), cm_etc)
        if err != nil {
            reqLogger.Error(err, "failed to create a new ConfigMap", "ConfigMap.Namespace", cm_etc.Namespace, "ConfigMap.Name", cm_etc.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get ironic-etc ConfigMap")
        return reconcile.Result{}, err
    }

    cm_dhcp_found := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "dhcp-bin", Namespace: instance.Namespace}, cm_dhcp_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new configmap
        cm_dhcp, _ := helpers.GetDHCPConfigMap(instance.Namespace)
        reqLogger.Info("Creating a new dhcp-bin configmap", "ConfigMap.Namespace", cm_dhcp.Namespace, "ConfigMap.Name", cm_dhcp.Name)
        err = r.client.Create(context.TODO(), cm_dhcp)
        if err != nil {
            reqLogger.Error(err, "failed to create a new ConfigMap", "ConfigMap.Namespace", cm_dhcp.Namespace, "ConfigMap.Name", cm_dhcp.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get dhcp-bin ConfigMap")
        return reconcile.Result{}, err
    }

    cm_dhcp_etc_found := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "dhcp-etc", Namespace: instance.Namespace}, cm_dhcp_etc_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new configmap
        cm_dhcp_etc, _ := helpers.GetDHCPEtcConfigMap(instance.Namespace)
        reqLogger.Info("Creating a new dhcp-etc configmap", "ConfigMap.Namespace", cm_dhcp_etc.Namespace, "ConfigMap.Name", cm_dhcp_etc.Name)
        err = r.client.Create(context.TODO(), cm_dhcp_etc)
        if err != nil {
            reqLogger.Error(err, "failed to create a new ConfigMap", "ConfigMap.Namespace", cm_dhcp_etc.Namespace, "ConfigMap.Name", cm_dhcp_etc.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get dhcp-etc ConfigMap")
        return reconcile.Result{}, err
    }

    // retrieve entries in configmap for images
    cm_images := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "images", Namespace: instance.Namespace}, cm_images)

    // create init jobs
    job_init_found := &batchv1.Job{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-db-init", Namespace: instance.Namespace}, job_init_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new db init job
        job_init := helpers.GetDbInitJob(instance.Namespace, cm_images.Data)
        reqLogger.Info("Creating a new ironic-db-init job", "Job.Namespace", job_init.Namespace, "Job.Name", job_init.Name)
        err = r.client.Create(context.TODO(), job_init)
        if err != nil {
            reqLogger.Error(err, "failed to create a new Job", "Job.Namespace", job_init.Namespace, "Job.Name", job_init.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get db-init job")
        return reconcile.Result{}, err
    }

    job_db_sync_found := &batchv1.Job{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-db-sync", Namespace: instance.Namespace}, job_db_sync_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new db sync job
        job_db_sync := helpers.GetDbSyncJob(instance.Namespace, cm_images.Data)
        reqLogger.Info("Creating a new ironic-db-sync job", "Job.Namespace", job_db_sync.Namespace, "Job.Name", job_db_sync.Name)
        err = r.client.Create(context.TODO(), job_db_sync)
        if err != nil {
            reqLogger.Error(err, "failed to create a new Job", "Job.Namespace", job_db_sync.Namespace, "Job.Name", job_db_sync.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get db-sync job")
        return reconcile.Result{}, err
    }

    // deploy DHCP only if needed
    dhcp_settings := &corev1.ConfigMap{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "dhcp-settings", Namespace: instance.Namespace}, dhcp_settings)
    external_dhcp, _ := strconv.ParseBool(dhcp_settings.Data["USE_EXTERNAL_DHCP"])
    if (! external_dhcp) {
        // Check if the dhcp service already exists, if not create a new one
        dhcp_service_found := &corev1.Service{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: "dhcp-server", Namespace: instance.Namespace}, dhcp_service_found)
        if err != nil && errors.IsNotFound(err) {
            // Define a new dhcp service
            dhcp_service := helpers.GetDHCPService(instance.Namespace)
            reqLogger.Info("Creating a new DHCP service", "Service.Namespace", dhcp_service.Namespace, "StatefulSet.Name", dhcp_service.Name)
            err = r.client.Create(context.TODO(), dhcp_service)
            if err != nil {
                reqLogger.Error(err, "failed to create new DHCP Service", "Service.Namespace", dhcp_service.Namespace, "Service.Name", dhcp_service.Name)
                return reconcile.Result{}, err
            }
        } else if err != nil {
            reqLogger.Error(err, "failed to get dhcp service")
            return reconcile.Result{}, err
        }
        // check if the DHCP deployment already exists, if not create a new one
        dhcp_deployment_found := &appsv1.Deployment{}
        err = r.client.Get(context.TODO(), types.NamespacedName{Name: "dhcp-server", Namespace: instance.Namespace}, dhcp_deployment_found)
        if err != nil && errors.IsNotFound(err) {
            // Define a new dhcp deployment
            dhcp_deployment := helpers.GetDHCPDeployment(instance.Namespace, cm_images.Data)
            reqLogger.Info("Creating a new DHCP deployment", "Deployment.Namespace", dhcp_deployment.Namespace, "Deployment.Name", dhcp_deployment.Name)
            err = r.client.Create(context.TODO(), dhcp_deployment)
            if err != nil {
                reqLogger.Error(err, "failed to create new DHCP Deployment", "Deployment.Namespace", dhcp_deployment.Namespace, "Deployment.Name", dhcp_deployment.Name)
                return reconcile.Result{}, err
            }
        } else if err != nil {
            reqLogger.Error(err, "failed to get dhcp deployment")
            return reconcile.Result{}, err
        }
    } else {
        reqLogger.Info("Skipping DHCP creation, as an external one will be used")
    }

    // Check if the deployment already exists, if not create a new one
    found := &appsv1.Deployment{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
    if err != nil && errors.IsNotFound(err) {
        // Define a new deployment
        dep := helpers.GetDeploymentForIronic(instance.Name, instance.Namespace, cm_images.Data)
        controllerutil.SetControllerReference(instance, dep, r.scheme)
        reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
        err = r.client.Create(context.TODO(), dep)
        if err != nil {
            reqLogger.Error(err, "failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            return reconcile.Result{}, err
        }
        // Deployment created successfully - return and requeue
        return reconcile.Result{Requeue: true}, nil
    } else if err != nil {
        reqLogger.Error(err, "failed to get Deployment")
        return reconcile.Result{}, err
    }

    // Ensure the deployment size is the same as the spec
    size := instance.Spec.Size
    if *found.Spec.Replicas != size {
        found.Spec.Replicas = &size
        err = r.client.Update(context.TODO(), found)
        if err != nil {
            reqLogger.Error(err, "failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
            return reconcile.Result{}, err
        }
        // Spec updated - return and requeue
        return reconcile.Result{Requeue: true}, nil
    }

    // Check if the service already exists, if not create a new one
    found_srv := &corev1.Service{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found_srv)
    if err != nil && errors.IsNotFound(err) {
        // Define a new service
        srv := helpers.GetServiceForIronicApi(instance.Name, instance.Namespace)

        reqLogger.Info("Creating a new Service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
        err = r.client.Create(context.TODO(), srv)
        if err != nil {
            reqLogger.Error(err, "failed to create new Service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
            return reconcile.Result{}, err
        }
        // Service created successfully - return and requeue
        return reconcile.Result{Requeue: true}, nil
    } else if err != nil {
        reqLogger.Error(err, "failed to get Service")
        return reconcile.Result{}, err
    }

    // Update the Ironic status with the pod names
    // List the pods for this ironic deployment
    podList := &corev1.PodList{}
    labelSelector := labels.SelectorFromSet(helpers.GetLabelsForIronic(instance.Name))
    listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
    err = r.client.List(context.TODO(), listOps, podList)
    if err != nil {
        reqLogger.Error(err, "failed to list pods", "Ironic.Namespace", instance.Namespace, "Ironic.Name", instance.Name)
        return reconcile.Result{}, err
    }
    podNames := helpers.GetPodNames(podList.Items)

    // Update status.Nodes if needed
    if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
        instance.Status.Nodes = podNames
        err := r.client.Update(context.TODO(), instance)
        if err != nil {
            reqLogger.Error(err, "failed to update Ironic status")
            return reconcile.Result{}, err
        }
    }
    return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *ironicv1alpha1.Ironic) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
