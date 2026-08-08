// Package k8s wraps client-go to provide a simple, TUI-friendly interface for
// browsing Kubernetes clusters.
package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client holds the clientsets used to talk to a single Kubernetes cluster.
type Client struct {
	config     *rest.Config
	dyn        dynamic.Interface
	clientset  *kubernetes.Clientset
	metricsCli *metricsclient.Clientset
	discovery  discovery.DiscoveryInterface
	context    string
	namespace  string
}

// ContextInfo describes a context available in the user's kubeconfig.
type ContextInfo struct {
	Name    string
	Cluster string
	User    string
	Current bool
}

// DefaultKubeconfigPath returns the path to the default kubeconfig, honoring
// the KUBECONFIG environment variable.
func DefaultKubeconfigPath() string {
	if p := os.Getenv("KUBECONFIG"); p != "" {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// Contexts loads all contexts from the kubeconfig.
func Contexts(kubeconfig string) ([]ContextInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	raw, err := loadingRules.Load()
	if err != nil {
		return nil, err
	}
	infos := make([]ContextInfo, 0, len(raw.Contexts))
	for name, ctx := range raw.Contexts {
		infos = append(infos, ContextInfo{
			Name:    name,
			Cluster: ctx.Cluster,
			User:    ctx.AuthInfo,
			Current: name == raw.CurrentContext,
		})
	}
	return infos, nil
}

// New builds a Client from a kubeconfig path and an optional context name.
// An empty context name uses the kubeconfig's current context.
func New(kubeconfig, context string) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		overrides.CurrentContext = context
	}
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	cfg, err := cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}
	if cfg.QPS == 0 {
		cfg.QPS = 50
	}
	if cfg.Burst == 0 {
		cfg.Burst = 100
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building clientset: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building dynamic client: %w", err)
	}
	metricsCli, err := metricsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building metrics client: %w", err)
	}

	raw, err := cc.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("loading raw kubeconfig: %w", err)
	}
	cur := raw.CurrentContext
	if context != "" {
		cur = context
	}
	ns := ""
	if c, ok := raw.Contexts[cur]; ok {
		ns = c.Namespace
	}

	return &Client{
		config:     cfg,
		dyn:        dyn,
		clientset:  clientset,
		metricsCli: metricsCli,
		discovery:  clientset.Discovery(),
		context:    cur,
		namespace:  ns,
	}, nil
}

// Context returns the active context name.
func (c *Client) Context() string { return c.context }

// Namespace returns the namespace associated with the active context in the
// kubeconfig, or "" when the context has none.
func (c *Client) Namespace() string { return c.namespace }

// Dynamic returns the dynamic client.
func (c *Client) Dynamic() dynamic.Interface { return c.dyn }

// Clientset returns the typed clientset.
func (c *Client) Clientset() *kubernetes.Clientset { return c.clientset }

// Metrics returns the metrics clientset.
func (c *Client) Metrics() *metricsclient.Clientset { return c.metricsCli }

// Config returns the underlying REST config.
func (c *Client) Config() *rest.Config { return c.config }

// Namespaces lists all namespaces in the cluster.
func (c *Client) Namespaces(ctx context.Context) ([]string, error) {
	ns, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ns.Items))
	for _, n := range ns.Items {
		names = append(names, n.Name)
	}
	return names, nil
}
