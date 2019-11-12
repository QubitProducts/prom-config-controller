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

package fake

import (
	v1beta1 "github.com/QubitProducts/prom-config-controller/pkg/apis/config/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeScrapes implements ScrapeInterface
type FakeScrapes struct {
	Fake *FakeConfigV1beta1
	ns   string
}

var scrapesResource = schema.GroupVersionResource{Group: "config.prometheus.io", Version: "v1beta1", Resource: "scrapes"}

var scrapesKind = schema.GroupVersionKind{Group: "config.prometheus.io", Version: "v1beta1", Kind: "Scrape"}

// Get takes name of the scrape, and returns the corresponding scrape object, and an error if there is any.
func (c *FakeScrapes) Get(name string, options v1.GetOptions) (result *v1beta1.Scrape, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(scrapesResource, c.ns, name), &v1beta1.Scrape{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Scrape), err
}

// List takes label and field selectors, and returns the list of Scrapes that match those selectors.
func (c *FakeScrapes) List(opts v1.ListOptions) (result *v1beta1.ScrapeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(scrapesResource, scrapesKind, c.ns, opts), &v1beta1.ScrapeList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ScrapeList{ListMeta: obj.(*v1beta1.ScrapeList).ListMeta}
	for _, item := range obj.(*v1beta1.ScrapeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested scrapes.
func (c *FakeScrapes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(scrapesResource, c.ns, opts))

}

// Create takes the representation of a scrape and creates it.  Returns the server's representation of the scrape, and an error, if there is any.
func (c *FakeScrapes) Create(scrape *v1beta1.Scrape) (result *v1beta1.Scrape, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(scrapesResource, c.ns, scrape), &v1beta1.Scrape{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Scrape), err
}

// Update takes the representation of a scrape and updates it. Returns the server's representation of the scrape, and an error, if there is any.
func (c *FakeScrapes) Update(scrape *v1beta1.Scrape) (result *v1beta1.Scrape, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(scrapesResource, c.ns, scrape), &v1beta1.Scrape{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Scrape), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeScrapes) UpdateStatus(scrape *v1beta1.Scrape) (*v1beta1.Scrape, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(scrapesResource, "status", c.ns, scrape), &v1beta1.Scrape{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Scrape), err
}

// Delete takes name of the scrape and deletes it. Returns an error if one occurs.
func (c *FakeScrapes) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(scrapesResource, c.ns, name), &v1beta1.Scrape{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeScrapes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(scrapesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.ScrapeList{})
	return err
}

// Patch applies the patch and returns the patched scrape.
func (c *FakeScrapes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Scrape, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(scrapesResource, c.ns, name, pt, data, subresources...), &v1beta1.Scrape{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Scrape), err
}
