package main

import (
	"bytes"
	"flag"
	"fmt"
	stdlog "log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"

	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/rest"
	"k8s.io/client-go/1.4/tools/clientcmd"
)

var (
	// Version of the service.
	Version string
	// BuildTime is the time the service was build.
	BuildTime string
	// BuildTag is the tag with which the service was build.
	BuildTag string

	kubeconfig = flag.String("kubeconfig", "", "Specifies a kubeconfig file (for usage of this tool from outside the cluster)")
	service    = flag.String("service", "", "Name of service endpoints to watch")
	namespace  = flag.String("namespace", "default", "Kubernetes namespace")

	logger log.Logger
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	// Setup logging
	logger = log.NewJSONLogger(os.Stdout)
	logger = log.NewContext(logger).With("ts", log.DefaultTimestamp)
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	// Redirect stdout to logger
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.NewStdlibAdapter(logger))

	// Startup message
	if Version == "" {
		Version = "dev"
	}
	logger.Log(
		"msg", "Watcher starting",
		"kubeconfig", *kubeconfig,
		"service", *service,
		"namespace", namespace,
		"version", Version,
		"buildTime", BuildTime,
		"buildTag", BuildTag,
	)

	// Validate input
	if *service == "" {
		logger.Log("msg", "No service name specified")
		os.Exit(1)
	}

	// Create a cluster configuration
	var err error
	var cfg *rest.Config
	if *kubeconfig == "" {
		// Connect from within the cluster
		cfg, err = rest.InClusterConfig()
	} else {
		// Connect from outside the cluster
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	if err != nil {
		logger.Log("msg", "Cannot create cluster config", "err", err)
		os.Exit(1)
	}

	// Create a client from the config
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Log("msg", "Cannot create client", "err", err)
		os.Exit(1)
	}

	errc := make(chan error, 1)

	// One goroutine loops forever printing out the endpoints for the given service
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				printEndpoints(client, *service, *namespace)
			}
		}
	}()

	// Another goroutine waits for signals
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		errc <- nil
	}()

	// Wait for any goroutine to exit
	if err := <-errc; err != nil {
		logger.Log("msg", "Watcher stopped with error", "err", err)
		os.Exit(1)
	} else {
		logger.Log("msg", "Watcher stopped")
	}
}

// printEndpoints finds all endpoints of the service and prints to stdout.
func printEndpoints(client *kubernetes.Clientset, serviceName, namespace string) {
	endpoints, err := client.Core().Endpoints(namespace).Get(serviceName)
	if err != nil {
		logger.Log("msg", "Cannot find endpoints of service", "service", serviceName, "namespace", namespace, "err", err)
		return
	}

	// https://godoc.org/k8s.io/client-go/1.4/pkg/api/v1#EndpointSubset
	// describes that we need to calculate the Cartesian product of
	// Addresses x Ports here to get all endpoints.
	var addrs []string
	var ports []int
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			addrs = append(addrs, addr.IP) // addr.Hostname is blank
		}
		for _, port := range subset.Ports {
			ports = append(ports, int(port.Port))
		}
	}
	var hostports []string
	for _, addr := range addrs {
		for _, port := range ports {
			hostports = append(hostports, fmt.Sprintf("%s:%d", addr, port))
		}
	}

	// Print the endpoints of the service
	var buf bytes.Buffer
	buf.WriteString(serviceName)
	buf.WriteString(": ")
	for i, hostport := range hostports {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(hostport)
	}
	buf.WriteByte('\n')
	logger.Log("msg", buf.String())
}
