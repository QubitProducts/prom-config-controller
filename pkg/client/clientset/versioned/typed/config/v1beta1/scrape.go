/*
Copyright 2022 The Kubernetes sample-controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/QubitProducts/prom-config-controller/pkg/apis/config/v1beta1"
	scheme "github.com/QubitProducts/prom-config-controller/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ScrapesGetter has a method to return a ScrapeInterface.
// A group's client should implement this interface.
type ScrapesGetter interface {
	Scrapes(namespace string) ScrapeInterface
}

// ScrapeInterface has methods to work with Scrape resources.
type ScrapeInterface interface {
	Create(ctx context.Context, scrape *v1beta1.Scrape, opts v1.CreateOptions) (*v1beta1.Scrape, error)
	Update(ctx context.Context, scrape *v1beta1.Scrape, opts v1.UpdateOptions) (*v1beta1.Scrape, error)
	UpdateStatus(ctx context.Context, scrape *v1beta1.Scrape, opts v1.UpdateOptions) (*v1beta1.Scrape, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.Scrape, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.ScrapeList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Scrape, err error)
	ScrapeExpansion
}

// scrapes implements ScrapeInterface
type scrapes struct {
	client rest.Interface
	ns     string
}

// newScrapes returns a Scrapes
func newScrapes(c *ConfigV1beta1Client, namespace string) *scrapes {
	return &scrapes{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the scrape, and returns the corresponding scrape object, and an error if there is any.
func (c *scrapes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scrapes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Scrapes that match those selectors.
func (c *scrapes) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ScrapeList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.ScrapeList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scrapes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested scrapes.
func (c *scrapes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("scrapes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a scrape and creates it.  Returns the server's representation of the scrape, and an error, if there is any.
func (c *scrapes) Create(ctx context.Context, scrape *v1beta1.Scrape, opts v1.CreateOptions) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("scrapes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scrape).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a scrape and updates it. Returns the server's representation of the scrape, and an error, if there is any.
func (c *scrapes) Update(ctx context.Context, scrape *v1beta1.Scrape, opts v1.UpdateOptions) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scrapes").
		Name(scrape.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scrape).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *scrapes) UpdateStatus(ctx context.Context, scrape *v1beta1.Scrape, opts v1.UpdateOptions) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scrapes").
		Name(scrape.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scrape).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the scrape and deletes it. Returns an error if one occurs.
func (c *scrapes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scrapes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *scrapes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scrapes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched scrape.
func (c *scrapes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("scrapes").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
