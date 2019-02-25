# Gremlin Operator

Chaos test Kubernetes pods using https://www.gremlin.com. Break things on purpose.

This is an open-source Operator for scheduling attacks on pods within a Kubernetes cluster using CRD's.

Attacks are scheduled using a Cron format field in CRD's. This creates a Kubernetes native cronjob that you can view using `kubectl get cronjobs`.

When an attack starts this Operator automatically creates Gremlin pod(s) on the same node(s) as the target pods to directly attack pods on the same node.

**Note:** Attacks scheduled from the Gremlin Web UI are not used by this Operator. All configuration is via CRD. However, attack results will show up in the Gremlin Web UI.

# Installation

### Create Gremlin Secret

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

### Deploy the Gremlin Operator

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

You can find some example CRD's in `deploy/crds`. For example, to chaos test nginx pods in the cluster by killing the pods once every 24 hours at midnight you can apply the following CRD.

```yaml
apiVersion: gremlin.kubedex.com/v1alpha1
kind: Gremlin
metadata:
  name: nginx-process-killer-gremlin
spec:
  type: attack-container
  gremlin: process_killer
  interval: 60
  process: '^nginx'
  signal: -9
  newest: true
  exact: true
  kill_children: true
  labels:
    app: nginx
  container_filter: "n([a-z])inx"
  restart_on_failure: false
  schedule: "0 0 * * *"
```

Save this as `gremlin_v1alpha1_gremlin_cr_shutdown_nginx.yaml` and then `kubectl apply -f gremlin_v1alpha1_gremlin_cr_shutdown_nginx.yaml`.

Under `spec` you can **optionally** add the following fields to override settings per CRD. You should not need to set these.

```yaml
  config_override:
    team_id: ""
    team_private_key: ""
    team_certificate: ""
    team_secret: ""
    identifier: ""
    client_tags: ""
    config_service: ""
    config_region: ""
    config_public_ip: ""
    config_public_hostname: ""
    config_local_ip: ""
    config_local_hostname: ""
```

**Note:** To create an adhoc immediate attack leave the `schedule:` field empty.

The `labels:` field is mandatory and determines which pod(s) to attack. The `container_filter:` is optional and provides a way to directly attack certain containers within the pod(s). This supports [Golang regexes](https://regex-golang.appspot.com/assets/html/index.html).

### Supported Attacks

You can find commented examples for each attack under `deploy/crds`.

| Attack         | Description                                                                                    | 
|----------------|------------------------------------------------------------------------------------------------|
| Process Killer | Use this to kill pods                                                                          |
| Shutdown       | This will shutdown entire nodes that the targets are running on (killing all pods on the node) |
| CPU            | Increases CPU utilisation to 100% in the targeted pod(s) on the specified number of cores      |
| Disk           | Fills up the disk inside the pod container(s) by the specified percentage                      |
| I/O            | Saturates IO in the targeted pod(s)                                                            |
| Memory         | Increases memory in targeted pod(s) by the specified amount in mb, gb or percentage            |
| DNS            | Blackhole DNS in the targeted pod(s)                                                           |
| Latency        | Increase network latency in the targeted pod(s) by the specified number of ms                  |
| Packet Loss    | Triggers packet loss in the targeted pod(s)                                                    |
| Black Hole     | Drops all network traffic in targeted pod(s)                                                   |

See the [Gremlin docs](https://help.gremlin.com/attacks/) for more information about the attacks.

**Note:** The time travel attack won't work on Kubernetes.

# Development setup

See `docs/development.md`.
