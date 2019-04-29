/*
Copyright 2018 The Kubernetes sample-controller Authors.

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

// FakeRuleGroups implements RuleGroupInterface
type FakeRuleGroups struct {
	Fake *FakeConfigV1beta1
	ns   string
}

var rulegroupsResource = schema.GroupVersionResource{Group: "config.prometheus.io", Version: "v1beta1", Resource: "rulegroups"}

var rulegroupsKind = schema.GroupVersionKind{Group: "config.prometheus.io", Version: "v1beta1", Kind: "RuleGroup"}

// Get takes name of the ruleGroup, and returns the corresponding ruleGroup object, and an error if there is any.
func (c *FakeRuleGroups) Get(name string, options v1.GetOptions) (result *v1beta1.RuleGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(rulegroupsResource, c.ns, name), &v1beta1.RuleGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.RuleGroup), err
}

// List takes label and field selectors, and returns the list of RuleGroups that match those selectors.
func (c *FakeRuleGroups) List(opts v1.ListOptions) (result *v1beta1.RuleGroupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(rulegroupsResource, rulegroupsKind, c.ns, opts), &v1beta1.RuleGroupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.RuleGroupList{}
	for _, item := range obj.(*v1beta1.RuleGroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ruleGroups.
func (c *FakeRuleGroups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(rulegroupsResource, c.ns, opts))

}

// Create takes the representation of a ruleGroup and creates it.  Returns the server's representation of the ruleGroup, and an error, if there is any.
func (c *FakeRuleGroups) Create(ruleGroup *v1beta1.RuleGroup) (result *v1beta1.RuleGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(rulegroupsResource, c.ns, ruleGroup), &v1beta1.RuleGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.RuleGroup), err
}

// Update takes the representation of a ruleGroup and updates it. Returns the server's representation of the ruleGroup, and an error, if there is any.
func (c *FakeRuleGroups) Update(ruleGroup *v1beta1.RuleGroup) (result *v1beta1.RuleGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(rulegroupsResource, c.ns, ruleGroup), &v1beta1.RuleGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.RuleGroup), err
}

// Delete takes name of the ruleGroup and deletes it. Returns an error if one occurs.
func (c *FakeRuleGroups) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(rulegroupsResource, c.ns, name), &v1beta1.RuleGroup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRuleGroups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(rulegroupsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.RuleGroupList{})
	return err
}

// Patch applies the patch and returns the patched ruleGroup.
func (c *FakeRuleGroups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.RuleGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(rulegroupsResource, c.ns, name, data, subresources...), &v1beta1.RuleGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.RuleGroup), err
}