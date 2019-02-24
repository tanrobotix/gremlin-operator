package gremlin

import (
	"fmt"
	"os"

	gremlinv1alpha1 "github.com/Kubedex/gremlin-operator/pkg/apis/gremlin/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	v1 "k8s.io/api/core/v1"
)

const (
	// GremlinTeamID is the Team ID (required for authentication)
	GremlinTeamID = "GREMLIN_TEAM_ID"

	// GremlinIdentifier custom name to assign to this client
	// (default is the host’s IP address)
	GremlinIdentifier = "GREMLIN_IDENTIFIER"

	// GremlinClientTags Comma-separated list of custom tags to
	// assign to this client
	// (e.g. GREMLIN_CLIENT_TAGS="zone=us-east1,role=mysql,foo=bar")
	GremlinClientTags = "GREMLIN_CLIENT_TAGS"

	// GremlinConfigService is service or group tag
	GremlinConfigService = "GREMLIN_CONFIG_SERVICE"

	// GremlinConfigRegion is region or datacenter
	GremlinConfigRegion = "GREMLIN_CONFIG_REGION"

	// GremlinConfigZone is the Availability Zone
	GremlinConfigZone = "GREMLIN_CONFIG_ZONE"

	// GremlinConfigPublicIP is the public IP address
	GremlinConfigPublicIP = "GREMLIN_CONFIG_PUBLIC_IP"

	// GremlinConfigPublicHostname is the public hostname
	GremlinConfigPublicHostname = "GREMLIN_CONFIG_PUBLIC_HOSTNAME"

	// GremlinConfigLocalIP is the internal IP address
	GremlinConfigLocalIP = "GREMLIN_CONFIG_LOCAL_IP"

	// GremlinConfigLocalHostname is the internal hostname
	GremlinConfigLocalHostname = "GREMLIN_CONFIG_LOCAL_HOSTNAME"

	// GremlinTeamCertificate is kubernetes secret name
	// default gremlin-cert
	GremlinTeamCertificate = "GREMLIN_TEAM_CERTIFICATE"

	// GremlinTeamCertificateSecretKey is the key of the certificate secret to select from
	// default gremlin.cert
	GremlinTeamCertificateSecretKey = "GREMLIN_TEAM_CERTIFICATE_SECRET_KEY"

	// GremlinTeamKeySecretKey is the key of the key secret to select from
	// default gremlin.key
	GremlinTeamKeySecretKey = "GREMLIN_TEAM_KEY_SECRET_KEY"
)

func buildArgs(cr *gremlinv1alpha1.Gremlin, containerID string) []string {
	spec := cr.Spec
	args := []string{}

	// build args for attack-container type
	if cr.Spec.Type == "attack-container" {
		args = append(args, "attack-container", containerID)

		// length is common parameter except for shutdown Gremlin
		if spec.Length > 0 && spec.Gremlin != "shutdown" {
			args = append(args, getArg(map[string]interface{}{"--length": spec.Length})...)
		}

		switch cr.Spec.Gremlin {
		case "cpu":
			args = append(args, getArg(map[string]interface{}{"--cores": spec.Cores})...)
		case "disk":
			args = append(args, getArg(map[string]interface{}{
				"--dir":        spec.Dir,
				"--workers":    spec.Workers,
				"--block-size": spec.BlockSize,
				"--percent":    spec.Percent,
			})...)
		case "dns":
			args = append(args, getArg(map[string]interface{}{
				"--device":      spec.Device,
				"--ip_address":  spec.IPAddress,
				"--ip_protocol": spec.IPProtocol,
			})...)
		case "io":
			args = append(args, getArg(map[string]interface{}{
				"--dir":         spec.Dir,
				"--mode":        spec.Mode,
				"--block-size":  spec.BlockSize,
				"--block-count": spec.BlockCount,
			})...)
		case "latency":
			args = append(args, getArg(map[string]interface{}{
				"--ms":          spec.Ms,
				"--device":      spec.Device,
				"--egress_port": spec.EgressPort,
				"--src_port":    spec.SrcPort,
				"--ipaddress":   spec.IPAddress,
				"--hostname":    spec.Hostname,
				"--ipprotocol":  spec.IPProtocol,
			})...)
		case "memory":
			args = append(args, getArg(map[string]interface{}{
				"--ms":        spec.Ms,
				"--gigabytes": spec.GigaBytes,
				"--megabytes": spec.MegaBytes,
				"--percent":   spec.Percent,
			})...)
		case "packet_loss":
			args = append(args, getArg(map[string]interface{}{
				"--corrupt":     spec.Corrupt,
				"--device":      spec.Device,
				"--egress_port": spec.EgressPort,
				"--src_port":    spec.SrcPort,
				"--ipaddress":   spec.IPAddress,
				"--hostname":    spec.Hostname,
				"--ipprotocol":  spec.IPProtocol,
			})...)
		case "process_killer":
			args = append(args, getArg(map[string]interface{}{
				"--interval":      spec.Interval,
				"--process":       spec.Process,
				"--signal":        spec.Signal,
				"--group":         spec.Group,
				"--user":          spec.User,
				"--newest":        spec.Newest,
				"--oldest":        spec.Oldest,
				"--exact":         spec.Exact,
				"--kill_children": spec.KillChildren,
				"--full":          spec.Full,
			})...)
		case "shutdown":
			args = append(args, getArg(map[string]interface{}{
				"--delay":  spec.Delay,
				"--reboot": spec.Reboot,
			})...)
		}
	}

	return args
}

func getArg(m map[string]interface{}) []string {
	subArg := []string{}
	for k, v := range m {
		switch v.(type) {
		case uint:
			cst := v.(uint)
			if cst > 0 {
				subArg = append(subArg, k, fmt.Sprint(cst))
			}
		case int:
			cst := v.(int)
			if cst > 0 {
				subArg = append(subArg, k, fmt.Sprint(cst))
			}
		case string:
			cst := v.(string)
			if len(cst) > 0 {
				subArg = append(subArg, k, cst)
			}
		case bool:
			subArg = append(subArg, k, fmt.Sprint(v))
		}
	}
	return subArg
}

func getEnv(env string, def string, override string) string {
	// return override regardless
	if len(override) > 0 {
		return override
	}
	// lookup environemnt and if value not found return default else return value
	v, found := os.LookupEnv(env)
	if !found {
		return def
	}
	return v
}

func buildEnv(cr *gremlinv1alpha1.Gremlin) []v1.EnvVar {
	optional := false
	// TODO: fill the overrides with spec configuration
	env := []v1.EnvVar{
		{
			Name:  "GREMLIN_TEAM_ID",
			Value: getEnv(GremlinTeamID, "", cr.Spec.TeamID),
		},
		{
			Name: "GREMLIN_TEAM_CERTIFICATE_OR_FILE",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: getEnv(GremlinTeamCertificate, "gremlin-cert", cr.Spec.Config.TeamCertificate),
					},
					Key:      getEnv(GremlinTeamCertificateSecretKey, "gremlin.cert", cr.Spec.Config.TeamCertificateSecretKey),
					Optional: &optional,
				},
			},
		},
		{
			Name: "GREMLIN_TEAM_PRIVATE_KEY_OR_FILE",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: getEnv(GremlinTeamCertificate, "gremlin-cert", cr.Spec.Config.TeamCertificate),
					},
					Key:      getEnv(GremlinTeamKeySecretKey, "gremlin.key", cr.Spec.Config.TeamKeySecretKey),
					Optional: &optional,
				},
			},
		},
	}

	return env
}

func getCronJobSpec(cr *gremlinv1alpha1.Gremlin, container string, containerID string, namespace string, node string) batchv2alpha1.CronJobSpec {
	return batchv2alpha1.CronJobSpec{
		Schedule:          cr.Spec.Schedule,
		ConcurrencyPolicy: batchv2alpha1.ForbidConcurrent,
		JobTemplate: batchv2alpha1.JobTemplateSpec{
			Spec: getBatchJobSpec(cr, container, containerID, namespace, node),
		},
	}
}

func getBatchJobSpec(cr *gremlinv1alpha1.Gremlin, container string, containerID string, namespace string, node string) batchv1.JobSpec {
	// set restart policy
	restart := v1.RestartPolicyNever
	if cr.Spec.RestartOnFailure {
		restart = v1.RestartPolicyOnFailure
	}

	return batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  cr.Name + container + "-job-container",
						Image: "gremlin/gremlin",
						Args:  buildArgs(cr, containerID),
						SecurityContext: &v1.SecurityContext{
							Capabilities: &v1.Capabilities{
								Add: []v1.Capability{"NET_ADMIN", "SYS_BOOT", "SYS_TIME", "KILL"},
							},
						},
						Env: buildEnv(cr),
						VolumeMounts: []v1.VolumeMount{
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
						},
					},
				},
				RestartPolicy: restart,
				// set the exact node we want to run this attack
				NodeName:    node,
				HostNetwork: true,
				HostPID:     true,
				Volumes: []v1.Volume{
					{
						Name: "docker-sock",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/var/run/docker.sock",
							},
						},
					},
					{
						Name: "gremlin-state",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/var/lib/gremlin",
							},
						},
					},
					{
						Name: "gremlin-logs",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/var/log/gremlin",
							},
						},
					},
				},
			},
		},
	}
}
