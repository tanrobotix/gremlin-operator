# gremlin-operator

Chaos test Kubernetes pods using https://www.gremlin.com

This is an open-source Operator for scheduling attacks on pods within a Kubernetes cluster using CRD's.

Attacks are scheduled using a Cron format field in CRD's. This creates a Kubernetes native cronjob that you can view using `kubectl get cronjobs`.

When an attack starts this Operator automatically injects a Gremlin container into the pod for the lifecycle of the attack.

**Note:** Attacks scheduled from the Gremlin Web UI are not used by this Operator. All configuration is via CRD. However, attack results will show up in the Gremlin Web UI.

# Installation

### Create Gremlin secret

If you do not already have your certificates locally, you can download them by going the teams page in the Gremlin Web IUI and selecting the team for which you’d like to install the client. From there you can select ‘Download’ to download the current certificate, or ‘Create New’ if you have not yet created your client certificates.

When you download your certificate files, they will have a name like YOUR_TEAM_NAME-client.priv_key.pem and YOUR_TEAM_NAME-client.pub_cert.pem. Rename these files to gremlin.key and gremlin.cert respectively. Then create your secret as follows:

```sh
kubectl create secret generic gremlin-team-cert --from-file=./gremlin.cert --from-file=./gremlin.key
```

### Setup Service Account

```sh
kubectl create -f deploy/service_account.yaml
```

### Setup RBAC

```sh
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
```

### Setup the CRD

```sh
kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml
```

### Deploy the gremlin-operator

To configure the operator edit `deploy/operator.yaml` and modify `<your team id>`.

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gremlin-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: gremlin-operator
  template:
    metadata:
      labels:
        name: gremlin-operator
    spec:
      serviceAccountName: gremlin-operator
      containers:
        - name: gremlin-operator
          image: kubedex/gremlin-operator
          command:
          - gremlin-operator
          imagePullPolicy: Always
          env:
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: OPERATOR_NAME
              value: "gremlin-operator"
            - name: GREMLIN_TEAM_ID
              value: "<your team id>"
            - name: "GREMLIN_TEAM_CERTIFICATE"
              value: "gremlin-cert"
            - name: "GREMLIN_TEAM_CERTIFICATE_SECRET_KEY"
              value: "gremlin.cert"
            - name: "GREMLIN_TEAM_KEY_SECRET_KEY"
              value: "gremlin.key"
```

Then create the Operator.

```sh
kubectl create -f deploy/operator.yaml
```

### Create a GremlinService CRD

The default controller will watch for GremlinService objects and create a pod for each CRD.

You can find some example CRD's in `deploy/crds`. For example, to run the Shutdown Gremlin on nginx pods in the cluster once every 24 hours at midnight you can apply the following CRD.

```yaml
apiVersion: gremlin.kubedex.com/v1alpha1
kind: Gremlin
metadata:
  name: gremlin-shutdown-nginx
spec:
  type: attack-container
  gremlin: shutdown
  delay: 60
  reboot: true
  labels:
    app: nginx
  container_filter: "n([a-z])inx"
  restart_on_filaure: false
  schedule: "0 0 * * *"
```

Save this as `gremlin_v1alpha1_gremlin_cr_shutdown_nginx.yaml` and then `kubectl apply -f gremlin_v1alpha1_gremlin_cr_shutdown_nginx.yaml`.

**Note:** To create an adhoc immediate attack leave the `schedule:` field empty.


# Development setup

## Prerequisites
---

- [dep][dep_tool] version v0.5.0+.
- [git][git_tool]
- [go][go_tool] version v1.10+.
- [docker][docker_tool] version 17.03+.
- [kubectl][kubectl_tool] version v1.11.0+.
- Access to a kubernetes v.1.11.0+ cluster (kind or minikube)

First, checkout and install the operator-sdk CLI:

```sh
$ mkdir -p $GOPATH/src/github.com/gremlin-framework
$ cd $GOPATH/src/github.com/operator-framework
$ git clone https://github.com/operator-framework/operator-sdk
$ cd operator-sdk
$ git checkout master
$ make dep
$ make install
```

Create and deploy an gremlin-operator using the SDK CLI:

```sh
# Clone an gremlin-operator project that defines the Gremlin CR.
$ mkdir -p $GOPATH/src/github.com/Kubedex/
$ cd $GOPATH/src/github.com/Kubedex/
$ git clone git@github.com:Kubedex/gremlin-operator.git
$ cd gremlin-operator

# Add a new API for the custom resource AppService
$ operator-sdk add api --api-version=gremlin.kubedex.io/v1alpha1 --kind=AppService

# Add a new controller that watches for AppService
$ operator-sdk add controller --api-version=gremlin.kubedex.io/v1alpha1 --kind=AppService

# Build and push the gremlin-operator image to a public registry such as docker.io
$ operator-sdk build kubedex/gremlin-operator
$ docker push kubedex/gremlin-operator

# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|kubedex/gremlin-operator|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|kubedex/gremlin-operator|g' deploy/operator.yaml

# create gremlin secrets
$ kubectl create secret generic gremlin-team-cert --from-file=./gremlin.cert --from-file=./gremlin.key

# Setup Service Account
$ kubectl create -f deploy/service_account.yaml
# Setup RBAC
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
# Setup the CRD
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml
# Deploy the gremlin-operator
$ kubectl create -f deploy/operator.yaml

# Create an GremlinService CR
# The default controller will watch for GremlinService objects and create a pod for each CR
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_cpu.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_disk.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_dns.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_io.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_latency.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_memory.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_packet_loss.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_process_killer.yaml
$ kubectl create -f deploy/crds/gremlin_v1alpha1_gremlin_cr_shutdown.yaml

# verify CR is created
$ kubectl get gremlins.gremlin.kubedex.com
NAME              AGE
example-gremlin   32s

# Verify that a job is created
$ kubectl get pod -l app=example-gremlinservice
NAME                     READY     STATUS    RESTARTS   AGE
example-appservice-pod   1/1       Running   0          1m

# Test the new Resource Type
$ kubectl describe gremlinservice example-gremlinservice
Name:         example-gremlinservice
Namespace:    myproject
Labels:       <none>
Annotations:  <none>
API Version:  gremlin.kubedex.io/v1alpha1
Kind:         AppService
Metadata:
  Cluster Name:        
  Creation Timestamp:  2018-12-17T21:18:43Z
  Generation:          1
  Resource Version:    248412
  Self Link:           /apis/gremlin.kubedex.io/v1alpha1/namespaces/gremlin/gremlinservices/example-gremlinservice
  UID:                 554f301f-0241-11e9-b551-080027c7d133
Spec:
  Size:  3

# Cleanup
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_cpu.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_disk.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_dns.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_io.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_latency.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_memory.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_packet_loss.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_process_killer.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr_shutdown.yaml

$ kubectl delete -f deploy/operator.yaml
$ kubectl delete -f deploy/role.yaml
$ kubectl delete -f deploy/role_binding.yaml
$ kubectl delete -f deploy/service_account.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml
```
