package main

import (
	"context"
	"fmt"
	"github.com/jp39/kubernetes-zfs-provisioner/pkg/provisioner"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v10/controller"
)

const (
	metricsAddrKey         = "METRICS_ADDR"
	metricsPortKey         = "METRICS_PORT"
	kubeConfigPathKey      = "KUBE_CONFIG_PATH"
	provisionerInstanceKey = "PROVISIONER_INSTANCE"
	parentDatasetKey       = "PARENT_DATASET"
)

type Settings struct {
	MetricsAddr         string
	MetricsPort         int
	KubeConfigPath      string
	ProvisionerInstance string
	parentDataset       string
}

var (
	// These will be populated by Goreleaser at build time
	version = "snapshot"
	commit  = "dirty"

	settings Settings
)

func main() {
	loadEnvironmentVariables()

	log := klog.NewKlogr()

	log.Info("Using configuration", "config", settings)

	config, err := clientcmd.BuildConfigFromFlags("", settings.KubeConfigPath)
	if err != nil {
		klog.Fatalf("Couldn't get in-cluster or kubectl config: %v", err)
	}

	// Retrieve config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create kubernetes client: %v", err)
	}

	log.Info("Connected to cluster", "host", config.Host)
	p, err := provisioner.NewZFSProvisioner(settings.ProvisionerInstance,
						settings.parentDataset, log)
	if err != nil {
		klog.Fatalf("Failed to create ZFS provisioner: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	pc := controller.NewProvisionController(
		log,
		clientset,
		settings.ProvisionerInstance,
		p,
		controller.MetricsAddress(settings.MetricsAddr),
		controller.MetricsPort(int32(settings.MetricsPort)),
	)

	log.Info("Starting provisioner", "version", version, "commit", commit)
	pc.Run(context.Background())
}

func loadEnvironmentVariables() {
	prefix := "ZFS_"

	defaults := map[string]string{
		metricsPortKey:         "8080",
		metricsAddrKey:         "0.0.0.0",
		kubeConfigPathKey:      "",
		provisionerInstanceKey: "pv.kubernetes.io/zfs",
		parentDatasetKey:       "tank/kubernetes",
	}

	for key, _ := range defaults {
		value, found := os.LookupEnv(fmt.Sprintf("%s%s", prefix, key))
		if found {
			defaults[key] = value
		}
	}
	settings = Settings{
		MetricsAddr:         defaults[metricsAddrKey],
		MetricsPort:         parseInt(defaults[metricsPortKey]),
		KubeConfigPath:      defaults[kubeConfigPathKey],
		ProvisionerInstance: defaults[provisionerInstanceKey],
		parentDataset:       defaults[parentDatasetKey],
	}
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		klog.Fatalf("Failed to convert metrics port to integer: %v", err)
	}
	return i
}
