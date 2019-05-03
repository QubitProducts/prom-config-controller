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
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"text/template"
	"time"

	promconfig "github.com/QubitProducts/prom-config-controller/internal/prom2"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	yaml "gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	configV1beta1 "github.com/QubitProducts/prom-config-controller/pkg/apis/config/v1beta1"
	clientset "github.com/QubitProducts/prom-config-controller/pkg/client/clientset/versioned"
	configScheme "github.com/QubitProducts/prom-config-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/QubitProducts/prom-config-controller/pkg/client/informers/externalversions"
	listers "github.com/QubitProducts/prom-config-controller/pkg/client/listers/config/v1beta1"
)

const controllerAgentName = "prom-config-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a conf is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a conf fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// ErrResourceInvalid is used as part of the Event 'reason' when a conf fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceInvalid = "ErrResourceInvalid"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by conf"
	// MessageResourceSynced is the message used for an Event fired when a conf
	// is synced successfully
	MessageResourceSynced = "conf synced successfully"
)

// ControllerConfig describes the controller config
type ControllerConfig struct {
	CACert      []byte
	ServiceName string
	ServiceNS   string

	Namespace string
	Selector  labels.Selector

	RuleConfigMapNS  string
	RuleConfigMap    string
	RuleConfigMapKey string
	RuleFile         string

	ConfigSecretNS  string
	ConfigSecret    string
	ConfigSecretKey string
	ConfigFile      string
	ConfigTemplate  *template.Template
}

// Controller describes the controller implementation for conf resources
type Controller struct {
	Reloader
	ControllerConfig

	kubeclientset kubernetes.Interface
	confclientset clientset.Interface

	rulesLister listers.RuleGroupLister
	rulesSynced cache.InformerSynced

	scrapesLister listers.ScrapeLister
	scrapesSynced cache.InformerSynced

	rulesWorkqueue   workqueue.RateLimitingInterface
	scrapesWorkqueue workqueue.RateLimitingInterface
	recorder         record.EventRecorder

	clusterLister clusterLister
}

// NewController returns a new sample controller
func NewController(
	cfg ControllerConfig,
	kubeclientset kubernetes.Interface,
	confclientset clientset.Interface,
	confInformerFactory informers.SharedInformerFactory,
	reloader Reloader,
	clusterLister clusterLister,
) *Controller {

	rulesInformer := confInformerFactory.Config().V1beta1().RuleGroups()
	scrapesInformer := confInformerFactory.Config().V1beta1().Scrapes()

	configScheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		Reloader:         reloader,
		ControllerConfig: cfg,

		kubeclientset:    kubeclientset,
		confclientset:    confclientset,
		rulesLister:      rulesInformer.Lister(),
		rulesSynced:      rulesInformer.Informer().HasSynced,
		scrapesLister:    scrapesInformer.Lister(),
		scrapesSynced:    scrapesInformer.Informer().HasSynced,
		rulesWorkqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter(), "rules"),
		scrapesWorkqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter(), "scrapes"),
		recorder:         recorder,
		clusterLister:    clusterLister,
	}

	glog.Info("Setting up event handlers")
	rulesInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueuerule,
		DeleteFunc: controller.enqueuerule,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueuerule(new)
		},
	})

	scrapesInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueuescrape,
		DeleteFunc: controller.enqueuescrape,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueuescrape(new)
		},
	})

	return controller
}

// Run the controller
func (c *Controller) Run(stopCh <-chan struct{}) error {
	glog.Info("Starting Prometheus Config controller")

	defer runtime.HandleCrash()
	defer c.rulesWorkqueue.ShutDown()
	defer c.scrapesWorkqueue.ShutDown()

	glog.Info("self registering validation webhook")
	if err := c.selfRegistration(); err != nil {
		glog.Errorf("registering webhook failed, %v", err)
		return errors.Wrap(err, "registering webhook failed")
	}

	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.rulesSynced, c.scrapesSynced); !ok {
		glog.Errorf("failed waiting for cache sync")
		return fmt.Errorf("caches did not sync")
	}

	confChanged_, err := c.syncConfigHandler()
	if err != nil {
		glog.Errorf("rintial config sync failed, %v", err)
	}

	rulesChanged, err := c.syncRuleHandler()
	if err != nil {
		glog.Errorf("intial rules sync failed, %v", err)
	}

	if (confChanged_ || rulesChanged) && c.Reloader != nil {
		glog.Infof("the configuration has changed, issuing reload")
		c.Reloader.Reload()
	}

	go wait.Until(c.runRulesWorker, time.Second, stopCh)
	go wait.Until(c.runConfigWorker, time.Second, stopCh)

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runRulesWorker() {
	processRule := makeProcessNextWorkItem(c.rulesWorkqueue, c.syncRuleHandler, c.Reloader)
	for processRule() {
	}

	glog.Info("rules worker stopped")
}

func (c *Controller) runConfigWorker() {
	processScrape := makeProcessNextWorkItem(c.scrapesWorkqueue, c.syncConfigHandler, c.Reloader)
	for processScrape() {
	}

	glog.Info("scrapes worker stopped")
}

func makeProcessNextWorkItem(
	workqueue workqueue.RateLimitingInterface,
	sync func() (bool, error),
	reloader Reloader) func() bool {
	return func() bool {
		obj, shutdown := workqueue.Get()

		if shutdown {
			return false
		}

		// We wrap this block in a func so we can defer c.workqueue.Done.
		err := func(obj interface{}) error {
			defer workqueue.Done(obj)
			var key string
			var ok bool
			if key, ok = obj.(string); !ok {
				workqueue.Forget(obj)
				runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
				return nil
			}
			reload, err := sync()
			if err != nil {
				glog.Infof("error syncing config, %v", err)
				return fmt.Errorf("error syncing %s, %v", key, err.Error())
			}
			workqueue.Forget(obj)
			if reload && reloader != nil {
				glog.Infof("the configuration has changed, issuing reload")
				reloader.Reload()
			}
			glog.V(2).Infof("Successfully synced '%s'", key)
			return nil
		}(obj)

		if err != nil {
			runtime.HandleError(err)
			return true
		}

		return true
	}
}

func (c *Controller) syncRuleHandler() (bool, error) {
	groupKeys := []string{}
	groups := map[string]*rulefmt.RuleGroup{}
	rr, err := c.rulesLister.RuleGroups(c.Namespace).List(c.Selector)
	if err != nil {
		return false, errors.Wrap(err, "listing rules")
	}

	for _, r := range rr {
		var rerrs []error
		var key string
		if key, err = cache.MetaNamespaceKeyFunc(r); err != nil {
			runtime.HandleError(err)
			continue
		}

		var res *rulefmt.RuleGroup
		res, rerrs = convertRuleGroup(r.GetName(), r)

		//		err = c.updateconftatus(r, err, rerrs)
		//		if err != nil {
		//			c.recorder.Event(r, corev1.EventTypeWarning, ErrResourceInvalid, MessageResourceSynced)
		//			continue
		//		}
		for _, err := range rerrs {
			glog.Infof("rule error in %v: %v", key, err)
		}

		groups[key] = res
		groupKeys = append(groupKeys, key)
	}

	final := &rulefmt.RuleGroups{}
	sort.Strings(groupKeys)
	for _, k := range groupKeys {
		final.Groups = append(final.Groups, *groups[k])
	}

	bs, err := yaml.Marshal(final)
	if err != nil {
		return false, errors.Wrap(err, "rendering rules yaml")
	}

	cmUpdated, err := c.updateConfigMap(c.RuleConfigMap, c.RuleConfigMapKey, c.RuleConfigMapNS, bs)
	if err != nil {
		return false, errors.Wrap(err, "udpate rules configmap")
	}

	fileChanged, err := updateFile(c.RuleFile, bs)
	if err != nil {
		return cmUpdated, errors.Wrap(err, "update rules file")
	}

	return cmUpdated || fileChanged, nil
}

func (c *Controller) updateConfigMap(name, key, namespace string, bs []byte) (bool, error) {
	if name == "" || key == "" {
		return false, nil
	}

	oldcm, err := c.kubeclientset.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		oldcm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string]string{},
		}
		oldcm, err = c.kubeclientset.CoreV1().ConfigMaps(namespace).Create(oldcm)
		if err != nil {
			return false, err
		}
	}

	if err != nil {
		return false, err
	}

	if bytes.Compare([]byte(oldcm.Data[key]), bs) == 0 {
		return false, nil
	}

	glog.Infof("configmap %s/%s[%s] changed, updating", namespace, name, key)

	newcm := oldcm.DeepCopy()
	if newcm.Data == nil {
		newcm.Data = make(map[string]string)
	}

	newcm.Data[key] = string(bs)
	_, err = c.kubeclientset.CoreV1().ConfigMaps(namespace).Update(newcm)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Controller) syncConfigHandler() (bool, error) {
	var configStr string
	if c.ConfigTemplate != nil {
		templateData := struct {
			Clusters []*cluster
		}{}
		var err error
		if c.clusterLister != nil {
			templateData.Clusters, err = c.clusterLister(context.TODO())
		}
		if err != nil {
			glog.Infof("listing clusters failed. %v", err)
		}
		glog.V(2).Infof("template:\n%v", c.ConfigTemplate)
		baseCfg := &bytes.Buffer{}
		if err := c.ConfigTemplate.Execute(baseCfg, templateData); err != nil {
			glog.V(2).Infof("error rendering template, %v", err)
			return false, errors.Wrap(err, "rendering config template")
		}
		configStr = baseCfg.String()
		glog.V(2).Infof("base config template output:\n%s", configStr)
	}

	basePromCfg, err := promconfig.Load(configStr)
	if err != nil {
		return false, errors.Wrap(err, "checking config template result")
	}

	scrapeKeys := []string{}
	scrapes := map[string]*promconfig.ScrapeConfig{}
	ss, err := c.scrapesLister.Scrapes(c.Namespace).List(c.Selector)
	if err != nil {
		return false, err
	}

	for _, s := range ss {
		var key string
		if key, err = cache.MetaNamespaceKeyFunc(s); err != nil {
			runtime.HandleError(err)
			continue
		}

		ps, err := convertScrape(s.GetName(), s)
		if err != nil {
			return false, errors.Wrap(err, "convert scrape config")
		}

		scrapes[key] = ps
		scrapeKeys = append(scrapeKeys, key)
	}

	sort.Strings(scrapeKeys)
	var scrapeList []*promconfig.ScrapeConfig
	for _, k := range scrapeKeys {
		scrapeList = append(scrapeList, scrapes[k])
	}

	basePromCfg.ScrapeConfigs = append(basePromCfg.ScrapeConfigs, scrapeList...)

	bs, err := yaml.Marshal(basePromCfg)
	if err != nil {
		return false, errors.Wrap(err, "convert config")
	}

	secUpdated, err := c.updateSecret(c.ConfigSecret, c.ConfigSecretKey, c.ConfigSecretNS, bs)
	if err != nil {
		return false, errors.Wrap(err, "udpate config secret")
	}

	fileChanged, err := updateFile(c.ConfigFile, bs)
	if err != nil {
		return secUpdated, errors.Wrap(err, "update config file")
	}

	return secUpdated || fileChanged, nil
}

func (c *Controller) updateSecret(name, key, namespace string, bs []byte) (bool, error) {
	if name == "" || key == "" {
		return false, nil
	}

	oldsec, err := c.kubeclientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		oldsec = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string][]byte{},
		}
		oldsec, err = c.kubeclientset.CoreV1().Secrets(namespace).Create(oldsec)
		if err != nil {
			return false, err
		}
	}
	if err != nil {
		return false, err
	}

	if bytes.Compare(oldsec.Data[key], bs) == 0 {
		return false, nil
	}

	glog.Infof("secret %s/%s[%s] changed, updating", namespace, name, key)

	newsec := oldsec.DeepCopy()
	if newsec.Data == nil {
		newsec.Data = make(map[string][]byte)
	}

	newsec.Data[key] = bs
	_, err = c.kubeclientset.CoreV1().Secrets(namespace).Update(newsec)
	if err != nil {
		return false, errors.Wrap(err, "writing config secret")
	}

	return true, nil
}

func updateFile(fn string, bs []byte) (bool, error) {
	if fn == "" {
		return false, nil
	}

	obs, err := ioutil.ReadFile(fn)
	if err != nil && !os.IsNotExist(err) {
		glog.V(3).Infof("err reading file, %v", err)

		return false, err
	}

	osha := sha1.Sum(obs)
	nsha := sha1.Sum(bs)
	if bytes.Compare(osha[:], nsha[:]) == 0 {
		return false, nil
	}

	glog.V(1).Infof("fn: %s old: %s new: %s",
		fn,
		base64.StdEncoding.EncodeToString(osha[:]),
		base64.StdEncoding.EncodeToString(nsha[:]))

	glog.Infof("file %s changed, updating", fn)

	return true, errors.Wrap(ioutil.WriteFile(fn, bs, 0644), "writing config file")
}

/*
func (c *Controller) updateconftatus(conf *configV1beta1.RuleGroup, cerr error, errs []error) error {
	confCopy := conf.DeepCopy()
	if cerr != nil {
		confCopy.Status.Errors = append(confCopy.Status.Errors, cerr.Error())
	}
	for _, err := range errs {
		confCopy.Status.Errors = append(confCopy.Status.Errors, err.Error())
	}
	_, err := c.confclientset.Config().RuleGroups(confCopy.Namespace).UpdateStatus(confCopy)
	return err
}
*/

func (c *Controller) enqueuerule(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.rulesWorkqueue.AddRateLimited(key)
}

func (c *Controller) enqueuescrape(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.scrapesWorkqueue.AddRateLimited(key)
}

func convertRuleGroup(name string, conf *configV1beta1.RuleGroup) (*rulefmt.RuleGroup, []error) {
	var err error

	var interval model.Duration
	if conf.Spec.Interval != "" {
		interval, err = model.ParseDuration(conf.Spec.Interval)
	}
	if err != nil {
		return nil, []error{errors.Wrap(err, "invalid interval")}
	}

	var errs []error
	var rules []rulefmt.Rule
	for i, r := range conf.Spec.Rules {
		var rfor model.Duration
		if r.For != "" {
			rfor, err = model.ParseDuration(r.For)
		}
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "invalid for duration in rule %v", i))
			continue
		}
		rules = append(rules, rulefmt.Rule{
			Alert:       r.Alert,
			Expr:        r.Expr,
			Record:      r.Record,
			Labels:      r.Labels,
			Annotations: r.Annotations,
			For:         rfor,
		})
	}

	if errs != nil {
		return nil, errs
	}

	rg := rulefmt.RuleGroup{
		Name:     fmt.Sprintf("%s/%s", conf.Namespace, conf.Name),
		Interval: model.Duration(interval),
		Rules:    rules,
	}

	rgs := rulefmt.RuleGroups{Groups: []rulefmt.RuleGroup{rg}}
	return &rg, rgs.Validate()
}

func convertScrape(name string, conf *configV1beta1.Scrape) (*promconfig.ScrapeConfig, error) {
	var pcfg promconfig.ScrapeConfig
	err := yaml.Unmarshal([]byte(conf.Spec), &pcfg)
	if err != nil {
		return nil, err
	}
	pcfg.JobName = fmt.Sprintf("%s/%s", conf.Namespace, conf.Name)

	return &pcfg, nil
}
