// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/kthcloud/go-deploy/pkg/imp/kubevirt/kubevirt/scheme"
	"net/http"

	rest "k8s.io/client-go/rest"
	v1alpha1 "kubevirt.io/api/snapshot/v1alpha1"
)

type SnapshotV1alpha1Interface interface {
	RESTClient() rest.Interface
	VirtualMachineRestoresGetter
	VirtualMachineSnapshotsGetter
	VirtualMachineSnapshotContentsGetter
}

// SnapshotV1alpha1Client is used to interact with features provided by the snapshot.kubevirt.io group.
type SnapshotV1alpha1Client struct {
	restClient rest.Interface
}

func (c *SnapshotV1alpha1Client) VirtualMachineRestores(namespace string) VirtualMachineRestoreInterface {
	return newVirtualMachineRestores(c, namespace)
}

func (c *SnapshotV1alpha1Client) VirtualMachineSnapshots(namespace string) VirtualMachineSnapshotInterface {
	return newVirtualMachineSnapshots(c, namespace)
}

func (c *SnapshotV1alpha1Client) VirtualMachineSnapshotContents(namespace string) VirtualMachineSnapshotContentInterface {
	return newVirtualMachineSnapshotContents(c, namespace)
}

// NewForConfig creates a new SnapshotV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*SnapshotV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new SnapshotV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*SnapshotV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &SnapshotV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new SnapshotV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *SnapshotV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new SnapshotV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *SnapshotV1alpha1Client {
	return &SnapshotV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *SnapshotV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
