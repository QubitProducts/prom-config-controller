/*
Copyright 2019 The Kubernetes sample-controller Authors.

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
	Create(*v1beta1.Scrape) (*v1beta1.Scrape, error)
	Update(*v1beta1.Scrape) (*v1beta1.Scrape, error)
	UpdateStatus(*v1beta1.Scrape) (*v1beta1.Scrape, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.Scrape, error)
	List(opts v1.ListOptions) (*v1beta1.ScrapeList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Scrape, err error)
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
func (c *scrapes) Get(name string, options v1.GetOptions) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("scrapes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Scrapes that match those selectors.
func (c *scrapes) List(opts v1.ListOptions) (result *v1beta1.ScrapeList, err error) {
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
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested scrapes.
func (c *scrapes) Watch(opts v1.ListOptions) (watch.Interface, error) {
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
		Watch()
}

// Create takes the representation of a scrape and creates it.  Returns the server's representation of the scrape, and an error, if there is any.
func (c *scrapes) Create(scrape *v1beta1.Scrape) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("scrapes").
		Body(scrape).
		Do().
		Into(result)
	return
}

// Update takes the representation of a scrape and updates it. Returns the server's representation of the scrape, and an error, if there is any.
func (c *scrapes) Update(scrape *v1beta1.Scrape) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scrapes").
		Name(scrape.Name).
		Body(scrape).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *scrapes) UpdateStatus(scrape *v1beta1.Scrape) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("scrapes").
		Name(scrape.Name).
		SubResource("status").
		Body(scrape).
		Do().
		Into(result)
	return
}

// Delete takes name of the scrape and deletes it. Returns an error if one occurs.
func (c *scrapes) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scrapes").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *scrapes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("scrapes").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched scrape.
func (c *scrapes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Scrape, err error) {
	result = &v1beta1.Scrape{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("scrapes").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
