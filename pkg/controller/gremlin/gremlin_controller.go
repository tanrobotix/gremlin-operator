package gremlin

import (
	"context"
	"fmt"
	"strings"

	gremlinv1alpha1 "github.com/Kubedex/gremlin-operator/pkg/apis/gremlin/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gremlin")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Gremlin Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGremlin{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gremlin-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Gremlin
	err = c.Watch(&source.Kind{Type: &gremlinv1alpha1.Gremlin{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Gremlin
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gremlinv1alpha1.Gremlin{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGremlin{}

// ReconcileGremlin reconciles a Gremlin object
type ReconcileGremlin struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Gremlin object and makes changes based on the state read
// and what is in the Gremlin.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGremlin) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Gremlin")

	// Fetch the Gremlin instance
	instance := &gremlinv1alpha1.Gremlin{}
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

	// Get all the pods with annotations
	podList := &v1.PodList{}
	// Return all pods in the request namespace with a label of `gremlin.gremlin.kubedex.com/attack=<value>`
	opts := &client.ListOptions{}
	opts.SetLabelSelector(fmt.Sprintf("gremlin.gremlin.kubedex.com/enabled=%s", "true"))
	err = r.client.List(context.TODO(), opts, podList)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, pod := range podList.Items {
		// check the attack is for single container or multiple containers
		attackVector := "all"
		if av, ok := pod.Annotations["gremlin.gremlin.kubedex.com/container"]; ok {
			attackVector = av
			reqLogger.Info("Attack selected", "Pod", av)
		}
		reqLogger.Info("Attack vector selected", "Pod", attackVector)

		// if attack is for all the containers, schedule pods for attack
		if attackVector == "all" {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				// Replace docker:// from container id
				containerID := strings.Replace(containerStatus.ContainerID, "docker://", "", 1)
				// Create a k8s job to attack pod container
				job := newJobForAttack(instance, containerStatus.Name, containerID,
					pod.Namespace, pod.Spec.NodeName, instance.Spec.TeamID)

				reqLogger.Info("Sheduling attack", "Job.Name", job.Name, "Job.Container", containerStatus.Name, "ContainerID", containerID)

				// Set Gremlin instance as the owner and controller
				if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
					return reconcile.Result{}, err
				}

				reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
				err = r.client.Create(context.TODO(), job)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
		}
	}
	// job schedule is success
	return reconcile.Result{}, nil
}

// newJobForAttack returns a gremlin/gremlin job with the same name/namespace and node as the pos
func newJobForAttack(cr *gremlinv1alpha1.Gremlin, container string, containerID string, namespace string, node string, teamID string) *batchv1.Job {
	labels := map[string]string{
		"app": cr.Name,
	}
	volumeSecretDefaultMode := int32(420)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + container + "-job",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cr.Name + container + "-job-pod",
							Image: "gremlin/gremlin",
							Args:  []string{"attack-container", containerID, "cpu"},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{"NET_ADMIN", "SYS_BOOT", "SYS_TIME", "KILL"},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "GREMLIN_TEAM_ID",
									Value: teamID,
								},
								{
									Name:  "GREMLIN_TEAM_CERTIFICATE_OR_FILE",
									Value: "file:///var/lib/gremlin/cert/gremlin.cert",
								},
								{
									Name:  "GREMLIN_TEAM_PRIVATE_KEY_OR_FILE",
									Value: "file:///var/lib/gremlin/cert/gremlin.key",
								},
								{
									Name:  "GREMLIN_IDENTIFIER",
									Value: cr.Name + container + "-job-pod",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "docker-sock",
									MountPath: "/var/run/docker.sock",
								},
								{
									Name:      "gremlin-state",
									MountPath: "/var/lib/gremlin",
								},
								{
									Name:      "gremlin-logs",
									MountPath: "/var/log/gremlin",
								},
								{
									Name:      "gremlin-cert",
									MountPath: "/var/lib/gremlin/cert",
									ReadOnly:  true,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					// set the exact node we want to run this attack
					NodeName:    node,
					HostNetwork: true,
					HostPID:     true,
					Volumes: []corev1.Volume{
						{
							Name: "docker-sock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/docker.sock",
								},
							},
						},
						{
							Name: "gremlin-state",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log/gremlin",
								},
							},
						},
						{
							Name: "gremlin-logs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/log/gremlin",
								},
							},
						},
						{
							Name: "gremlin-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "gremlin-team-cert",
									DefaultMode: &volumeSecretDefaultMode,
								},
							},
						},
					},
				},
			},
		},
	}
}
