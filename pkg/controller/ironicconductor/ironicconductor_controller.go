package ironicconductor

import (
	"context"
    "reflect"

	ironicv1alpha1 "github.com/redhat-nfvpe/ironic-operator/pkg/apis/ironic/v1alpha1"
    helpers "github.com/redhat-nfvpe/ironic-operator/pkg/helpers"

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

    // create init jobs
    job_init_found := &batchv1.Job{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-db-init", Namespace: instance.Namespace}, job_init_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new db init job
        job_init := r.GetDbInitJob(instance.Namespace)
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
        job_db_sync := r.GetDbSyncJob(instance.Namespace)
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
    job_rabbit_init_found := &batchv1.Job{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: "ironic-rabbit-init", Namespace: instance.Namespace}, job_rabbit_init_found)
    if err != nil && errors.IsNotFound(err) {
        // define a new rabbit init job
        job_rabbit_init := r.GetRabbitInitJob(instance.Namespace)
        reqLogger.Info("Creating a new ironic-rabbit-init job", "Job.Namespace", job_rabbit_init.Namespace, "Job.Name", job_rabbit_init.Name)
        err = r.client.Create(context.TODO(), job_rabbit_init)
        if err != nil {
            reqLogger.Error(err, "failed to create a new Job", "Job.Namespace", job_rabbit_init.Namespace, "Job.Name", job_rabbit_init.Name)
            return reconcile.Result{}, err
        }
    } else if err != nil {
        reqLogger.Error(err, "failed to get rabbit-init job")
        return reconcile.Result{}, err
    }

    // Check if the deployment already exists, if not create a new one
    found := &appsv1.StatefulSet{}
    err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
    if err != nil && errors.IsNotFound(err) {
        // Define a new stateful set
        sta := r.statefulSetForIronicConductor(instance)
        reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", sta.Namespace, "StatefulSet.Name", sta.Name)
        err = r.client.Create(context.TODO(), sta)
        if err != nil {
            reqLogger.Error(err, "failed to create new StatefulSet", "StatefulSet.Namespace", sta.Namespace, "StatefulSet.Name", sta.Name)
            return reconcile.Result{}, err
        }
        // Stateful set created successfully - return and requeue
        return reconcile.Result{Requeue: true}, nil
    } else if err != nil {
        reqLogger.Error(err, "failed to get Stateful set")
        return reconcile.Result{}, err
    }

    // Ensure the deployment size is the same as the spec
    size := instance.Spec.Size
    if *found.Spec.Replicas != size {
        found.Spec.Replicas = &size
        err = r.client.Update(context.TODO(), found)
        if err != nil {
            reqLogger.Error(err, "failed to update Stateful set", "StatefulSEt.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
            return reconcile.Result{}, err
        }
        // Spec updated - return and requeue
        return reconcile.Result{Requeue: true}, nil
    }

    // Update the Ironic Conductor status with the pod names
    // List the pods for this ironic conductor's deployment
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

// statefulSetForIronicConductor returns a ironic-conductor StatefulSet object
func (r *ReconcileIronicConductor) statefulSetForIronicConductor(m *ironicv1alpha1.IronicConductor) *appsv1.StatefulSet {
    ls := labelsForIronicConductor(m.Name)
    replicas := m.Spec.Size

    var readMode int32 = 0444
    var execMode int32 = 0555
    var rootUser int64 = 0
    var privTrue bool = true

    // Set IronicConductor instance as the owner and controller
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    sta := &appsv1.StatefulSet{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "apps/v1",
            Kind:       "StatefulSet",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      m.Name,
            Namespace: m.Namespace,
        },
        Spec: appsv1.StatefulSetSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: ls,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: ls,
                },
                Spec: corev1.PodSpec {
                    NodeSelector: node_selector,
                    SecurityContext: &corev1.PodSecurityContext {
                        RunAsUser: &rootUser,
                    },
                    HostNetwork: true,
                    HostIPC: true,
                    DNSPolicy: "ClusterFirstWithHostNet",
                    InitContainers: []corev1.Container{
                        {
                            Name: "init",
                            Image: "quay.io/stackanetes/kubernetes-entrypoint:v0.3.1",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []corev1.EnvVar {
                                {
                                    Name: "POD_NAME",
                                    ValueFrom: &corev1.EnvVarSource {
                                        FieldRef: &corev1.ObjectFieldSelector {
                                            APIVersion: "v1",
                                            FieldPath: "metadata.name",
                                        },
                                    },
                                },
                                {
                                    Name: "NAMESPACE",
                                    ValueFrom: &corev1.EnvVarSource {
                                        FieldRef: &corev1.ObjectFieldSelector {
                                            APIVersion: "v1",
                                            FieldPath: "metadata.namespace",
                                        },
                                    },
                                },
                                {
                                    Name: "PATH",
                                    Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/",
                                },
                                {
                                    Name: "DEPENDENCY_SERVICE",
                                    Value: "default:ironic-api,default:rabbitmq",
                                },
                                {
                                    Name: "DEPENDENCY_JOBS",
                                    Value: "ironic-db-sync,ironic-rabbit-init",
                                },
                                {
                                    Name: "COMMAND",
                                    Value: "echo done",
                                },
                            },
                            Command: []string { "kubernetes-entrypoint" },
                        },
                        {
                            Name: "ironic-conductor-pxe-init",
                            Image: "docker.io/tripleorocky/centos-binary-ironic-pxe:current-tripleo",
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/ironic-conductor-pxe-init.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-pxe-init.sh",
                                    SubPath: "ironic-conductor-pxe-init.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-init",
                            Image: "quay.io/yrobla/tripleorocky-centos-binary-ironic-conductor",
                            ImagePullPolicy: "IfNotPresent",
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-init.sh",
                                    SubPath: "ironic-conductor-init.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-http-init",
                            Image: "quay.io/yrobla/tripleorocky-centos-binary-ironic-conductor",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []corev1.EnvVar {
                                {
                                    Name: "PXE_NIC",
                                    ValueFrom: &corev1.EnvVarSource {
                                        ConfigMapKeyRef: &corev1.ConfigMapKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "pxe-settings",
                                            },
                                            Key: "PXE_NIC",
                                        },
                                    },
                                },
                            },
                            Command: []string { "/tmp/ironic-conductor-http-init.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-http-init.sh",
                                    SubPath: "ironic-conductor-http-init.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/nginx/nginx.conf",
                                    SubPath: "nginx.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                            },
                        },
                    },
                    Containers: []corev1.Container {
                        {
                            Name: "ironic-conductor",
                            Image: "quay.io/yrobla/tripleorocky-centos-binary-ironic-conductor",
                            ImagePullPolicy: "IfNotPresent",
                            SecurityContext: &corev1.SecurityContext {
                                Privileged: &privTrue,
                                RunAsUser: &rootUser,
                            },
                            Command: []string { "/tmp/ironic-conductor.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor.sh",
                                    SubPath: "ironic-conductor.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                                {
                                    Name: "pod-var-cache-ironic",
                                    MountPath: "/var/cache/ironic",
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-pxe",
                            Image: "docker.io/tripleorocky/centos-binary-ironic-pxe:current-tripleo",
                            ImagePullPolicy: "IfNotPresent",
                            SecurityContext: &corev1.SecurityContext {
                                Privileged: &privTrue,
                                RunAsUser: &rootUser,
                            },
                            Env: []corev1.EnvVar {
                                {
                                    Name: "PXE_NIC",
                                    ValueFrom: &corev1.EnvVarSource {
                                        ConfigMapKeyRef: &corev1.ConfigMapKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "pxe-settings",
                                            },
                                            Key: "PXE_NIC",
                                        },
                                    },
                               },
                            },
                            Command: []string { "/tmp/ironic-conductor-pxe.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-pxe.sh",
                                    SubPath: "ironic-conductor-pxe.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "ironic-etc",
                                    MountPath: "/tftp-map-file",
                                    SubPath: "tftp-map-file",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                            Ports: []corev1.ContainerPort {
                                {
                                    ContainerPort: 69,
                                    HostPort: 69,
                                    Protocol: "UDP",
                                },
                            },
                        },
                        {
                            Name: "ironic-conductor-http",
                            Image: "docker.io/nginx:1.13.3",
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/ironic-conductor-http.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "ironic-bin",
                                    MountPath: "/tmp/ironic-conductor-http.sh",
                                    SubPath: "ironic-conductor-http.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "pod-shared",
                                    MountPath: "/tmp/pod-shared",
                                },
                                {
                                    Name: "pod-data",
                                    MountPath: "/var/lib/pod_data",
                                },
                            },
                            Ports: []corev1.ContainerPort {
                                {
                                    ContainerPort: 8081,
                                    HostPort: 8081,
                                    Protocol: "TCP",
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume {
                        {
                            Name: "pod-shared",
                            VolumeSource: corev1.VolumeSource {
                                EmptyDir: &corev1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "pod-data",
                            VolumeSource: corev1.VolumeSource {
                                EmptyDir: &corev1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "pod-var-cache-ironic",
                            VolumeSource: corev1.VolumeSource {
                                EmptyDir: &corev1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "ironic-bin",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "ironic-etc",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    controllerutil.SetControllerReference(m, sta, r.scheme)
    return sta
}

func (r *ReconcileIronicConductor) GetDbInitJob(namespace string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var readMode int32 = 0444
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-db-init",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "db-init" },
                },
                Spec: corev1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    Containers: []corev1.Container {
                        {
                            Name: "ironic-db-init-0",
                            Image: "quay.io/yrobla/tripleorocky-centos-binary-ironic-api",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []corev1.EnvVar {
                                {
                                    Name: "ROOT_DB_HOST",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_HOST",
                                        },
                                    },
                                },
                                {
                                    Name: "ROOT_DB_USER",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_USER",
                                        },
                                    },
                                },
                                {
                                    Name: "ROOT_DB_PASSWORD",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "mysql-root-credentials",
                                            },
                                            Key: "ROOT_DB_PASSWORD",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_HOST",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_HOST",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_USER",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_USER",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_PASSWORD",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_PASSWORD",
                                        },
                                    },
                                },
                                {
                                    Name: "USER_DB_DATABASE",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-db-user",
                                            },
                                            Key: "DB_DATABASE",
                                        },
                                    },
                                },

                            },
                            Command: []string { "/tmp/db-init.py" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "db-init-py",
                                    MountPath: "/tmp/db-init.py",
                                    SubPath: "db-init.py",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "etc-service",
                                    MountPath: "/etc/ironic",
                                },
                                {
                                    Name: "db-init-conf",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "db-init-conf",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume {
                        {
                            Name: "etc-service",
                            VolumeSource: corev1.VolumeSource {
                                EmptyDir: &corev1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "db-init-py",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "db-init-conf",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
}

func (r *ReconcileIronicConductor) GetDbSyncJob(namespace string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var readMode int32 = 0444
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-db-sync",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "db-sync" },
                },
                Spec: corev1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    InitContainers: []corev1.Container {
                        {
                            Name: "init",
                            Image: "quay.io/stackanetes/kubernetes-entrypoint:v0.3.1",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []corev1.EnvVar {
                                {
                                    Name: "PATH",
                                    Value: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/",
                                },
                                {
                                    Name: "DEPENDENCY_JOBS",
                                    Value: "ironic-db-init",
                                },
                                {
                                    Name: "COMMAND",
                                    Value: "echo done",
                                },
                            },
                            Command: []string { "kubernetes-entrypoint" },
                        },
                    },
                    Containers: []corev1.Container {
                        {
                            Name: "ironic-db-sync",
                            Image: "quay.io/yrobla/tripleorocky-centos-binary-ironic-api",
                            ImagePullPolicy: "IfNotPresent",
                            Command: []string { "/tmp/db-sync.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "db-sync-sh",
                                    MountPath: "/tmp/db-sync.sh",
                                    SubPath: "db-sync.sh",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "etc-service",
                                    MountPath: "/etc/ironic",
                                },
                                {
                                    Name: "db-sync-conf",
                                    MountPath: "/etc/ironic/ironic.conf",
                                    SubPath: "ironic.conf",
                                    ReadOnly: true,
                                },
                                {
                                    Name: "db-sync-conf",
                                    MountPath: "/etc/ironic/logging.conf",
                                    SubPath: "logging.conf",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume {
                        {
                            Name: "etc-service",
                            VolumeSource: corev1.VolumeSource {
                                EmptyDir: &corev1.EmptyDirVolumeSource {},
                            },
                        },
                        {
                            Name: "db-sync-sh",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                        {
                            Name: "db-sync-conf",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &readMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-etc",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
}

func (r *ReconcileIronicConductor) GetRabbitInitJob(namespace string) *batchv1.Job {
    node_selector := map[string]string{"ironic-control-plane": "enabled"}
    var execMode int32 = 0555

    job := &batchv1.Job{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "batch/v1",
            Kind:       "Job",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name:      "ironic-rabbit-init",
            Namespace: namespace,
        },
        Spec: batchv1.JobSpec {
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta {
                    Labels: map[string]string {"app": "ironic", "ironicapi_cr": "openstack-ironicapi", "component": "rabbit-init" },
                },
                Spec: corev1.PodSpec {
                    NodeSelector: node_selector,
                    RestartPolicy: "OnFailure",
                    Containers: []corev1.Container {
                        {
                            Name: "rabbit-init",
                            Image: "docker.io/rabbitmq:3.7-management",
                            ImagePullPolicy: "IfNotPresent",
                            Env: []corev1.EnvVar {
                                {
                                    Name: "RABBITMQ_ADMIN_CONNECTION",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-rabbitmq-admin",
                                            },
                                            Key: "RABBITMQ_CONNECTION",
                                        },
                                    },
                                },
                                {
                                    Name: "RABBITMQ_USER_CONNECTION",
                                    ValueFrom: &corev1.EnvVarSource {
                                        SecretKeyRef: &corev1.SecretKeySelector {
                                            LocalObjectReference: corev1.LocalObjectReference {
                                                Name: "ironic-rabbitmq-user",
                                            },
                                            Key: "RABBITMQ_CONNECTION",
                                        },
                                    },
                                },
                            },
                            Command: []string { "/tmp/rabbit-init.sh" },
                            VolumeMounts: []corev1.VolumeMount {
                                {
                                    Name: "rabbit-init-sh",
                                    MountPath: "/tmp/rabbit-init.sh",
                                    SubPath: "rabbit-init.sh",
                                    ReadOnly: true,
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume {
                        {
                            Name: "rabbit-init-sh",
                            VolumeSource: corev1.VolumeSource {
                                ConfigMap: &corev1.ConfigMapVolumeSource {
                                    DefaultMode: &execMode,
                                    LocalObjectReference: corev1.LocalObjectReference {
                                        Name: "ironic-bin",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return job
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
