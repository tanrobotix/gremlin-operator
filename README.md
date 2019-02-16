# gremlin-operator
Chaos test Kubernetes pods using https://www.gremlin.com

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
$ kubectl create -f deploy/crds/gremlin_v1alpha1_appservice_cr.yaml

# Verify that a pod is created
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
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_cr.yaml
$ kubectl delete -f deploy/operator.yaml
$ kubectl delete -f deploy/role.yaml
$ kubectl delete -f deploy/role_binding.yaml
$ kubectl delete -f deploy/service_account.yaml
$ kubectl delete -f deploy/crds/gremlin_v1alpha1_gremlin_crd.yaml
```