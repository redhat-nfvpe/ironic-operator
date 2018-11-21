package ironicconductor

import (
	"context"
    "reflect"

	ironicv1alpha1 "github.com/redhat-nfvpe/ironic-operator/pkg/apis/ironic/v1alpha1"
    helpers "github.com/redhat-nfvpe/ironic-operator/pkg/helpers"

    appsv1 "k8s.io/api/apps/v1"
    authv1 "k8s.io/api/rbac/v1"
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

var log = logf.Log.WithName("controller_ironicconductor")

// Add creates a new IronicConductor Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIronicConductor{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ironicconductor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IronicConductor
	err = c.Watch(&source.Kind{Type: &ironicv1alpha1.IronicConductor{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ironicv1alpha1.IronicConductor{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileIronicConductor{}

// ReconcileIronicConductor reconciles a IronicConductor object
type ReconcileIronicConductor struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a IronicConductor object and makes changes based on the state read
// and what is in the IronicConductor.Spec
func (r *ReconcileIronicConductor) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IronicConductor")

	// Fetch the IronicConductor instance
	instance := &ironicv1alpha1.IronicConductor{}
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
        cm_etc, _ := helpers.GetIronicEtcConfigMap(instance.Namespace)
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

    // Check if the service accounts, roles, etc... already exist, or create new
    // ones if needed
    sa_found := &corev1.ServiceAccount{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-conductor", Namespace: instance.Namespace}, sa_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new service account
        sa := r.ServiceAccountForIronicConductor(instance)
        reqLogger.Info("Creating a new ironic-conductor service account", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
        err = r.client.Create(context.TODO(), sa)
        if err != nil {
            reqLogger.Error(err, "failed to create a new ServiceAccount", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get ironic-conductor ServiceAccount")
        return reconcile.Result{}, err
    }

    rb_found := &authv1.RoleBinding{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-conductor", Namespace: instance.Namespace}, rb_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new role binding
        rb := r.RoleBindingForIronicConductor(instance)
        reqLogger.Info("Creating a new ironic-conductor role binding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
        err = r.client.Create(context.TODO(), rb)
        if err != nil {
            reqLogger.Error(err, "failed to create a new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get ironic-conductor RoleBinding")
        return reconcile.Result{}, err
    }

    // Check if the deployment already exists, if not create a new one
    found := &appsv1.Deployment{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
    if err != nil && errors.IsNotFound(err) {
        // Define a new deployment
        dep := r.deploymentForIronicConductor(instance)
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

    // Update the Ironic Conductor status with the pod names
    // List the pods for this ironic api's deployment
    podList := &corev1.PodList{}
    labelSelector := labels.SelectorFromSet(labelsForIronicConductor(instance.Name))
    listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
    err = r.client.List(context.TODO(), listOps, podList)
    if err != nil {
        reqLogger.Error(err, "failed to list pods", "IronicConductor.Namespace", instance.Namespace, "IronicConductor.Name", instance.Name)
        return reconcile.Result{}, err
    }
    podNames := getPodNames(podList.Items)

    // Update status.Nodes if needed
    if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
        instance.Status.Nodes = podNames
        err := r.client.Update(context.TODO(), instance)
        if err != nil {
            reqLogger.Error(err, "failed to update IronicConductor status")
            return reconcile.Result{}, err
        }
    }

    return reconcile.Result{}, nil
}

// deploymentForIronicConductor returns a ironic-conductor Deployment object
func (r *ReconcileIronicConductor) deploymentForIronicConductor(m *ironicv1alpha1.IronicConductor) *appsv1.Deployment {
    ls := labelsForIronicConductor(m.Name)
    replicas := m.Spec.Size

    var readMode int32 = 0444
    var execMode int32 = 0555

    // Set IronicConductor instance as the owner and controller
    dep := &appsv1.Deployment{

    }

    controllerutil.SetControllerReference(m, dep, r.scheme)
    return dep
}

// labelsForIronicConductor returns the labels for selecting the resources
// belonging to the given ironic conductor CR name.
func labelsForIronicConductor(name string) map[string]string {
    return map[string]string{"app": "ironic", "ironicconductor_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
        var podNames []string
        for _, pod := range pods {
                podNames = append(podNames, pod.Name)
        }
        return podNames
}

func (r *ReconcileIronicConductor) ServiceAccountForIronicConductor(m *ironicv1alpha1.IronicConductor) *corev1.ServiceAccount {
    sa := &corev1.ServiceAccount {
        TypeMeta: metav1.TypeMeta {
            APIVersion: "core/v1",
            Kind: "ServiceAccount",
        },
        ObjectMeta: metav1.ObjectMeta {
            Name: m.Name,
            Namespace: m.Namespace,
        },
    }
    return sa
}

func (r *ReconcileIronicConductor)RoleBindingForIronicConductor(m *ironicv1alpha1.IronicConductor) *authv1.RoleBinding {
    rb := &authv1.RoleBinding {
        TypeMeta: metav1.TypeMeta {
            APIVersion: "rbac/v1",
            Kind: "RoleBinding",
        },
        ObjectMeta: metav1.ObjectMeta {
            Name: m.Name,
            Namespace: m.Namespace,
        },
        RoleRef: authv1.RoleRef {
            APIGroup: "rbac.authorization.k8s.io",
            Kind: "Role",
            Name: "default-ironic-conductor",
        },
        Subjects: []authv1.Subject {
            {
                Kind: "ServiceAccount",
                Name: m.Name,
                Namespace: m.Namespace,
            },
        },
    }
    return rb
}

func (r *ReconcileIronicConductor)RoleForIronicConductor(m *ironicv1alpha1.IronicConductor) *authv1.Role {
    role := &authv1.Role {
        TypeMeta: metav1.TypeMeta {
            APIVersion: "rbac/v1",
            Kind: "Role",
        },
        ObjectMeta: metav1.ObjectMeta {
            Name: m.Name,
            Namespace: m.Namespace,
        },
        Rules: []authv1.PolicyRule {
            {
                APIGroups: []string { "", "extensions", "batch", "apps" },
                Verbs: []string { "Get", "List" },
                Resources: []string { "Services", "Endpoints", "Jobs", "Pods" },
            },
        },
    }
    return role
}
