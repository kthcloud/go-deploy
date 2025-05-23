// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	scheme "github.com/kthcloud/go-deploy/pkg/imp/kubevirt/kubevirt/scheme"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "kubevirt.io/api/core/v1"
)

// VirtualMachineInstanceReplicaSetsGetter has a method to return a VirtualMachineInstanceReplicaSetInterface.
// A group's client should implement this interface.
type VirtualMachineInstanceReplicaSetsGetter interface {
	VirtualMachineInstanceReplicaSets(namespace string) VirtualMachineInstanceReplicaSetInterface
}

// VirtualMachineInstanceReplicaSetInterface has methods to work with VirtualMachineInstanceReplicaSet resources.
type VirtualMachineInstanceReplicaSetInterface interface {
	Create(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.CreateOptions) (*v1.VirtualMachineInstanceReplicaSet, error)
	Update(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.UpdateOptions) (*v1.VirtualMachineInstanceReplicaSet, error)
	UpdateStatus(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.UpdateOptions) (*v1.VirtualMachineInstanceReplicaSet, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.VirtualMachineInstanceReplicaSet, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.VirtualMachineInstanceReplicaSetList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.VirtualMachineInstanceReplicaSet, err error)
	VirtualMachineInstanceReplicaSetExpansion
}

// virtualMachineInstanceReplicaSets implements VirtualMachineInstanceReplicaSetInterface
type virtualMachineInstanceReplicaSets struct {
	client rest.Interface
	ns     string
}

// newVirtualMachineInstanceReplicaSets returns a VirtualMachineInstanceReplicaSets
func newVirtualMachineInstanceReplicaSets(c *KubevirtV1Client, namespace string) *virtualMachineInstanceReplicaSets {
	return &virtualMachineInstanceReplicaSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the virtualMachineInstanceReplicaSet, and returns the corresponding virtualMachineInstanceReplicaSet object, and an error if there is any.
func (c *virtualMachineInstanceReplicaSets) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.VirtualMachineInstanceReplicaSet, err error) {
	result = &v1.VirtualMachineInstanceReplicaSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VirtualMachineInstanceReplicaSets that match those selectors.
func (c *virtualMachineInstanceReplicaSets) List(ctx context.Context, opts metav1.ListOptions) (result *v1.VirtualMachineInstanceReplicaSetList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.VirtualMachineInstanceReplicaSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested virtualMachineInstanceReplicaSets.
func (c *virtualMachineInstanceReplicaSets) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a virtualMachineInstanceReplicaSet and creates it.  Returns the server's representation of the virtualMachineInstanceReplicaSet, and an error, if there is any.
func (c *virtualMachineInstanceReplicaSets) Create(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.CreateOptions) (result *v1.VirtualMachineInstanceReplicaSet, err error) {
	result = &v1.VirtualMachineInstanceReplicaSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineInstanceReplicaSet).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a virtualMachineInstanceReplicaSet and updates it. Returns the server's representation of the virtualMachineInstanceReplicaSet, and an error, if there is any.
func (c *virtualMachineInstanceReplicaSets) Update(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.UpdateOptions) (result *v1.VirtualMachineInstanceReplicaSet, err error) {
	result = &v1.VirtualMachineInstanceReplicaSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		Name(virtualMachineInstanceReplicaSet.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineInstanceReplicaSet).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *virtualMachineInstanceReplicaSets) UpdateStatus(ctx context.Context, virtualMachineInstanceReplicaSet *v1.VirtualMachineInstanceReplicaSet, opts metav1.UpdateOptions) (result *v1.VirtualMachineInstanceReplicaSet, err error) {
	result = &v1.VirtualMachineInstanceReplicaSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		Name(virtualMachineInstanceReplicaSet.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(virtualMachineInstanceReplicaSet).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the virtualMachineInstanceReplicaSet and deletes it. Returns an error if one occurs.
func (c *virtualMachineInstanceReplicaSets) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *virtualMachineInstanceReplicaSets) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched virtualMachineInstanceReplicaSet.
func (c *virtualMachineInstanceReplicaSets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.VirtualMachineInstanceReplicaSet, err error) {
	result = &v1.VirtualMachineInstanceReplicaSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("virtualmachineinstancereplicasets").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
