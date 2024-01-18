package memcached

import (
	"fmt"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultPort         = 11211
	dataVolumeName      = "data"
	exporterDefaultPort = 9150
)

// MemcachedOptions is the options for the memcached container.
type MemcachedOptions struct {
	ConnLimit       int    `opt:"conn-limit"`
	ListenBacklog   int    `opt:"listen-backlog"`
	MaxItemSize     string `opt:"max-item-size"`
	MaxReqsPerEvent int    `opt:"max-reqs-per-event"`
	MemoryLimit     int    `opt:"memory-limit"`
	Port            int    `opt:"port"`
	Threads         int    `opt:"threads"`
	Verbose         bool   `opt:"verbose"`
	VeryVerbose     bool   `opt:"vv,single-hyphen"`

	// Extra options not included above.
	cmdopt.ExtraOpts
}

// MemcachedDeployment is the memcached deployment.
type MemcachedDeployment struct {
	Options *MemcachedOptions
	k8sutil.DeploymentGenericConfig
	ExporterImage    string
	ExporterImageTag string
}

// NewMemcachedStatefulSet returns a new memcached deployment.
func NewMemcached() *MemcachedDeployment {
	options := MemcachedOptions{}

	commonLabels := map[string]string{
		k8sutil.NameLabel:      "memcached",
		k8sutil.InstanceLabel:  "observatorium",
		k8sutil.PartOfLabel:    "observatorium",
		k8sutil.ComponentLabel: "memcached",
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     commonLabels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: commonLabels[k8sutil.InstanceLabel],
	}

	genericDeployment := k8sutil.DeploymentGenericConfig{
		Name:                          "memcached",
		Image:                         "docker.io/memcached",
		ImageTag:                      "latest",
		ImagePullPolicy:               corev1.PullIfNotPresent,
		CommonLabels:                  commonLabels,
		Replicas:                      1,
		Env:                           []corev1.EnvVar{},
		ContainerResources:            k8sutil.NewResourcesRequirements("500m", "3", "2Gi", "3Gi"),
		Affinity:                      k8sutil.NewAntiAffinity(nil, labelSelectors),
		EnableServiceMonitor:          true,
		TerminationGracePeriodSeconds: 120,
		SecurityContext:               k8sutil.GetDefaultSecurityContext(),
		ConfigMaps:                    map[string]map[string]string{},
		Secrets:                       map[string]map[string][]byte{},
	}

	return &MemcachedDeployment{
		Options:                 &options,
		DeploymentGenericConfig: genericDeployment,
		ExporterImage:           "quay.io/prometheus/memcached-exporter",
		ExporterImageTag:        "latest",
	}
}

// Manifests returns the manifests for the memcached deployment.
func (s *MemcachedDeployment) Manifests() k8sutil.ObjectMap {
	if s.EnableServiceMonitor {
		s.Sidecars = append(s.Sidecars, s.makeExporterContainer())
	}

	container := s.makeContainer()
	ret := k8sutil.ObjectMap{}
	ret.AddAll(s.GenerateObjectsDeployment(container))

	// Set headless service to get stable network ID.
	service := k8sutil.GetObject[*corev1.Service](ret, "")
	service.Spec.ClusterIP = corev1.ClusterIPNone

	return ret
}

func (s *MemcachedDeployment) makeContainer() *k8sutil.Container {
	if s.Options == nil {
		s.Options = &MemcachedOptions{}
	}

	httpPort := defaultPort
	if s.Options.Port != 0 {
		httpPort = s.Options.Port
	}

	ret := s.ToContainer()
	ret.Name = "memcached"
	ret.Args = cmdopt.GetOpts(s.Options)
	ret.Ports = []corev1.ContainerPort{
		{
			Name:          "client",
			ContainerPort: int32(httpPort),
			Protocol:      corev1.ProtocolTCP,
		},
	}
	ret.ServicePorts = []corev1.ServicePort{
		k8sutil.NewServicePort("client", httpPort, httpPort),
	}

	return ret
}

func (s *MemcachedDeployment) makeExporterContainer() *k8sutil.Container {
	return &k8sutil.Container{
		Name:            "memcached-exporter",
		Image:           s.ExporterImage,
		ImageTag:        s.ExporterImageTag,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       k8sutil.NewResourcesRequirements("50m", "200m", "50Mi", "200Mi"),
		Args: []string{
			fmt.Sprintf("--memcached.address=localhost:%d", s.Options.Port),
			fmt.Sprintf("--web.listen-address=:%d", exporterDefaultPort),
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "metrics",
				ContainerPort: exporterDefaultPort,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ServicePorts: []corev1.ServicePort{
			k8sutil.NewServicePort("metrics", exporterDefaultPort, exporterDefaultPort),
		},
		MonitorPorts: []monv1.Endpoint{
			{
				Port:           "metrics",
				RelabelConfigs: k8sutil.GetDefaultServiceMonitorRelabelConfig(),
			},
		},
	}
}