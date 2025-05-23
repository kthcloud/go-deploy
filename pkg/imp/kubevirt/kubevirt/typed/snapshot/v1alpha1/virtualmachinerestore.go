// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	scheme "github.com/kthcloud/go-deploy/pkg/imp/kubevirt/kubevirt/scheme"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "kubevirt.io/api/snapshot/v1alpha1"
)

// VirtualMachineRestoresGetter has a method to return a VirtualMachineRestoreInterface.
// A group's client should implement this interface.
type VirtualMachineRestoresGetter interface {
	VirtualMachineRestores(namespace string) VirtualMachineRestoreInterface
}

// VirtualMachineRestoreInterface has methods to work with VirtualMachineRestore resources.
type VirtualMachineRestoreInterface interface {
	Create(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.CreateOptions) (*v1alpha1.VirtualMachineRestore, error)
	Update(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.UpdateOptions) (*v1alpha1.VirtualMachineRestore, error)
	UpdateStatus(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.UpdateOptions) (*v1alpha1.VirtualMachineRestore, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.VirtualMachineRestore, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VirtualMachineRestoreList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtualMachineRestore, err error)
	VirtualMachineRestoreExpansion
}

// virtualMachineRestores implements VirtualMachineRestoreInterface
type virtualMachineRestores struct {
	client rest.Interface
	ns     string
}

// newVirtualMachineRestores returns a VirtualMachineRestores
func newVirtualMachineRestores(c *SnapshotV1alpha1Client, namespace string) *virtualMachineRestores {
	return &virtualMachineRestores{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the virtualMachineRestore, and returns the corresponding virtualMachineRestore object, and an error if there is any.
func (c *virtualMachineRestores) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VirtualMachineRestore, err error) {
	result = &v1alpha1.VirtualMachineRestore{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VirtualMachineRestores that match those selectors.
func (c *virtualMachineRestores) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VirtualMachineRestoreList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.VirtualMachineRestoreList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested virtualMachineRestores.
func (c *virtualMachineRestores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a virtualMachineRestore and creates it.  Returns the server's representation of the virtualMachineRestore, and an error, if there is any.
func (c *virtualMachineRestores) Create(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.CreateOptions) (result *v1alpha1.VirtualMachineRestore, err error) {
	result = &v1alpha1.VirtualMachineRestore{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineRestore).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a virtualMachineRestore and updates it. Returns the server's representation of the virtualMachineRestore, and an error, if there is any.
func (c *virtualMachineRestores) Update(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.UpdateOptions) (result *v1alpha1.VirtualMachineRestore, err error) {
	result = &v1alpha1.VirtualMachineRestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		Name(virtualMachineRestore.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineRestore).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *virtualMachineRestores) UpdateStatus(ctx context.Context, virtualMachineRestore *v1alpha1.VirtualMachineRestore, opts v1.UpdateOptions) (result *v1alpha1.VirtualMachineRestore, err error) {
	result = &v1alpha1.VirtualMachineRestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		Name(virtualMachineRestore.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineRestore).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the virtualMachineRestore and deletes it. Returns an error if one occurs.
func (c *virtualMachineRestores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *virtualMachineRestores) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched virtualMachineRestore.
func (c *virtualMachineRestores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtualMachineRestore, err error) {
	result = &v1alpha1.VirtualMachineRestore{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("virtualmachinerestores").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
