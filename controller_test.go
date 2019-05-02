/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	conf "github.com/QubitProducts/prom-config-controller/pkg/apis/config/v1beta1"
	"github.com/QubitProducts/prom-config-controller/pkg/client/clientset/versioned/fake"
	informers "github.com/QubitProducts/prom-config-controller/pkg/client/informers/externalversions"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

var testGroup = `
rules:
- record: something
  expr: 1 + 1
- record: something2
  expr: 1 + 1`

var testConfigMap = `groups:
- name: default/test
  rules:
  - record: something
    expr: 1 + 1
  - record: something2
    expr: 1 + 1
`

var testScrape = `job_name: extra-server
metrics_path: /metrics
scheme: http
gce_sd_configs:
- project: myproject
  zone: europe-west1-b
  filter: name eq mymonolith.*
  refresh_interval: 1m
  port: 1234
  tag_separator: ','
relabel_configs:
- source_labels: [__meta_gce_instance_name]
  separator: ;
  regex: (.*)
  target_label: instance
  replacement: $1
  action: replace`

var testSecret = `global:
  scrape_interval: 1m
  scrape_timeout: 10s
  evaluation_interval: 1m
scrape_configs:
- job_name: default/test
  metrics_path: /metrics
  scheme: http
  gce_sd_configs:
  - project: myproject
    zone: europe-west1-b
    filter: name eq mymonolith.*
    refresh_interval: 1m
    port: 1234
    tag_separator: ','
  relabel_configs:
  - source_labels: [__meta_gce_instance_name]
    separator: ;
    regex: (.*)
    target_label: instance
    replacement: $1
    action: replace
`

type fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset
	// Objects to put in the store.
	ruleGroupLister []*conf.RuleGroup
	scrapeLister    []*conf.Scrape
	// Actions expected to happen on the client.
	kubeactions []core.Action
	actions     []core.Action
	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newRuleGroup(name string, s string) *conf.RuleGroup {
	rg := conf.RuleGroupSpec{}
	err := yaml.Unmarshal([]byte(s), &rg)
	if err != nil {
		panic(err)
	}

	return &conf.RuleGroup{
		TypeMeta: metav1.TypeMeta{APIVersion: conf.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: rg,
	}
}

func newScrape(name string, s string) *conf.Scrape {
	return &conf.Scrape{
		TypeMeta: metav1.TypeMeta{APIVersion: conf.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: conf.ScrapeSpec(s),
	}
}

func newConfigMap(namespace, name, key, data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{key: data},
	}
}

func newSecret(namespace, name, key, data string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{key: []byte(data)},
	}
}

type testReloader struct {
	state bool
}

func (r *testReloader) Reload() {
	r.state = true
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

	reloader := testReloader{}

	sel, _ := labels.Parse("")
	cfg := ControllerConfig{
		Namespace:        "default",
		Selector:         sel,
		RuleConfigMapNS:  "default",
		RuleConfigMap:    "prom-config-controller",
		RuleConfigMapKey: "prom-config-controller.yaml",
		ConfigSecretNS:   "default",
		ConfigSecret:     "prom-config-controller",
		ConfigSecretKey:  "prom-config-controller.yaml",
	}
	c := NewController(cfg,
		f.kubeclient,
		f.client,
		i,
		&reloader,
		func(ctx context.Context) ([]*cluster, error) { return nil, nil })

	c.rulesSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, f := range f.ruleGroupLister {
		i.Config().V1beta1().RuleGroups().Informer().GetIndexer().Add(f)
	}

	for _, f := range f.scrapeLister {
		i.Config().V1beta1().Scrapes().Informer().GetIndexer().Add(f)
	}

	return c, i, k8sI
}

func (f *fixture) run(obj interface{}, t *testing.T) {
	f.runController(obj, true, false, t)
}

func (f *fixture) runExpectError(obj interface{}, t *testing.T) {
	f.runController(obj, true, true, t)
}

func (f *fixture) runController(obj interface{}, startInformers bool, expectError bool, t *testing.T) {
	c, i, k8sI := f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		i.Start(stopCh)
		k8sI.Start(stopCh)
	}

	switch obj.(type) {
	case *conf.RuleGroup:
		_, err := c.syncRuleHandler()
		if !expectError && err != nil {
			f.t.Errorf("error syncing rulegroup: %v", err)
		} else if expectError && err == nil {
			f.t.Error("expected error syncing rulegroup, got nil")
		}
	case *conf.Scrape:
		_, err := c.syncConfigHandler()
		if !expectError && err != nil {
			f.t.Errorf("error syncing scrape: %v", err)
		} else if expectError && err == nil {
			f.t.Error("expected error syncing scrape, got nil")
		}
	default:
	}

	actions := filterInformerActions(f.client.Actions())
	f.t.Logf("actions: %#v", actions)
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.kubeclient.Actions())
	f.t.Logf("k8s actions: %#v", k8sActions)
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected k8s actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected k8s actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateAction:
		e, _ := expected.(core.CreateAction)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expObject, object))
		}
	case core.UpdateAction:
		e, _ := expected.(core.UpdateAction)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expObject, object))
		}
	case core.PatchAction:
		e, _ := expected.(core.PatchAction)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, expPatch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintDiff(expPatch, patch))
		}
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if action.Matches("get", "configmaps") ||
			action.Matches("list", "configmaps") ||
			action.Matches("watch", "configmaps") ||
			action.Matches("get", "secrets") ||
			action.Matches("list", "secrets") ||
			action.Matches("watch", "secrets") ||
			action.Matches("get", "scrapes") ||
			action.Matches("list", "scrapes") ||
			action.Matches("watch", "scrapes") ||
			action.Matches("get", "rulegroups") ||
			action.Matches("list", "rulegroups") ||
			action.Matches("watch", "rulegroups") {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectCreateRuleGroupsAction(rs *conf.RuleGroup) {
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{
		Resource: "rulegroups",
		Group:    conf.SchemeGroupVersion.Group,
		Version:  conf.SchemeGroupVersion.Version,
	}, rs.Namespace, rs))
}

func (f *fixture) expectUpdateRuleGroupsAction(rs *conf.RuleGroup) {
	f.actions = append(f.actions, core.NewUpdateAction(schema.GroupVersionResource{
		Resource: "rulegroups",
		Group:    conf.SchemeGroupVersion.Group,
		Version:  conf.SchemeGroupVersion.Version,
	}, rs.Namespace, rs))
}

func (f *fixture) expectCreateConfigMapAction(cm *corev1.ConfigMap) {
	f.kubeactions = append(f.kubeactions, core.NewCreateAction(schema.GroupVersionResource{
		Resource: "configmaps",
		Group:    "",
		Version:  "v1",
	}, cm.Namespace, cm))
}

func (f *fixture) expectCreateSecretAction(s *corev1.Secret) {
	f.kubeactions = append(f.kubeactions, core.NewCreateAction(schema.GroupVersionResource{
		Resource: "secrets",
		Group:    "",
		Version:  "v1",
	}, s.Namespace, s))
}

func (f *fixture) expectUpdateConfigMapAction(cm *corev1.ConfigMap) {
	f.kubeactions = append(f.kubeactions, core.NewUpdateAction(schema.GroupVersionResource{
		Resource: "configmaps",
		Group:    "",
		Version:  "v1",
	}, cm.Namespace, cm))
}

func (f *fixture) expectUpdateSecretAction(s *corev1.Secret) {
	f.kubeactions = append(f.kubeactions, core.NewUpdateAction(schema.GroupVersionResource{
		Resource: "secrets",
		Group:    "",
		Version:  "v1",
	}, s.Namespace, s))
}

func (f *fixture) expectCreateScrapeAction(s *conf.Scrape) {
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{
		Resource: "scrapes",
		Group:    conf.SchemeGroupVersion.Group,
		Version:  conf.SchemeGroupVersion.Version,
	}, s.Namespace, s))
}

func (f *fixture) expectUpdateScrapeAction(s *conf.Scrape) {
	f.actions = append(f.actions, core.NewUpdateAction(schema.GroupVersionResource{
		Resource: "scrapes",
		Group:    conf.SchemeGroupVersion.Group,
		Version:  conf.SchemeGroupVersion.Version,
	}, s.Namespace, s))
}

func getKey(obj interface{}, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		t.Errorf("Unexpected error getting key for %v: %v", obj, err)
		return ""
	}
	return key
}

func TestCreatesRuleGroup(t *testing.T) {
	f := newFixture(t)
	rs := newRuleGroup("test", testGroup)
	cm := newConfigMap(
		"default",
		"prom-config-controller",
		"prom-config-controller.yaml",
		testConfigMap)
	f.ruleGroupLister = append(f.ruleGroupLister, rs)
	f.objects = append(f.objects, rs)
	f.kubeobjects = append(f.kubeobjects, cm)
	f.expectUpdateConfigMapAction(cm)

	f.run(rs, t)
}

func TestCreatesScrape(t *testing.T) {
	f := newFixture(t)
	scrape := newScrape("test", testScrape)
	s := newSecret(
		"default",
		"prom-config-controller",
		"prom-config-controller.yaml",
		testSecret)
	f.scrapeLister = append(f.scrapeLister, scrape)
	f.objects = append(f.objects, scrape)
	f.kubeobjects = append(f.kubeobjects, s)
	f.expectUpdateSecretAction(s)

	f.run(scrape, t)
}

/*
func TestDoNothing(t *testing.T) {
	f := newFixture(t)
	rs := newRuleGroup("test", testGroup)
	cmap := newConfigMap("test", testGroup)

	f.ruleGroupLister = append(f.ruleGroupLister, rs)
	f.objects = append(f.objects, rs)
	f.expectCreateRuleGroupsAction(cmap)

	f.run(getKey(rs, t))
}

func TestUpdateRuleGroup(t *testing.T) {
	f := newFixture(t)
	rs := newRuleGroup("test", testGroup)

	f.ruleGroupLister = append(f.ruleGroupLister, rs)
	f.objects = append(f.objects, rs)
	f.expectCreateRuleGroupsAction(rs)

	f.run(getKey(rs, t))
}
*/
