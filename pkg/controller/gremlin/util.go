package gremlin

import (
	"fmt"

	gremlinv1alpha1 "github.com/Kubedex/gremlin-operator/pkg/apis/gremlin/v1alpha1"
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

	return []string{}
}

func getArg(m map[string]interface{}) []string {
	subArg := []string{}
	for k, v := range m {
		switch v.(type) {
		case uint:
			cst := v.(uint)
			if cst > 0 {
				subArg = append([]string{k, fmt.Sprint(cst)})
			}
		case int:
			cst := v.(int)
			if cst > 0 {
				subArg = append([]string{k, fmt.Sprint(cst)})
			}
		case string:
			cst := v.(string)
			if cst != "" {
				subArg = append([]string{k, cst})
			}
		case bool:
			subArg = append([]string{k, fmt.Sprint(v)})
		}
	}
	return subArg
}
