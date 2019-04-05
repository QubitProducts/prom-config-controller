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
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/errgroup"

	gke "cloud.google.com/go/container/apiv1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/QubitProducts/prom-config-controller/pkg/client/clientset/versioned"
	informers "github.com/QubitProducts/prom-config-controller/pkg/client/informers/externalversions"
	"github.com/QubitProducts/prom-config-controller/pkg/signals"
)

var (
	addr       string
	masterURL  string
	kubeconfig string

	serviceNS   string
	serviceName string

	rulesMapNS   string
	rulesMapName string
	rulesMapKey  string
	rulesFile    string

	configTemplate string

	configSecNS   string
	configSecName string
	configSecKey  string
	configFile    string

	namespace string
	selector  string

	tlsKey  string
	tlsCA   string
	tlsCert string

	reloadScheme      string
	reloadHost        string
	reloadMethod      string
	reloadPath        string
	reloadEndpointsNS string
	reloadEndpoints   string
	reloadDelay       time.Duration
	reloadRetries     int

	gcpProject string
	gcpKeysDir string
)

func init() {
	flag.StringVar(&addr, "addr", ":8443", "Address to listen on.")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.StringVar(&namespace, "namespace", "", "namespace to watch for resources")
	flag.StringVar(&selector, "labels", "", "label selector for resources")
	flag.StringVar(&configTemplate, "config.template", "config.yaml.tmpl", "")
	flag.StringVar(&rulesMapNS, "rules.configmap.namespace", "infra", "")
	flag.StringVar(&rulesMapName, "rules.configmap.name", "prom-config-controller", "")
	flag.StringVar(&rulesMapKey, "rules.configmap.key", "rules.yaml", "")
	flag.StringVar(&rulesFile, "rules.file", "rules.yaml", "")
	flag.StringVar(&configSecNS, "config.secret.namespace", "infra", "")
	flag.StringVar(&configSecName, "config.secret.name", "prom-config-controller", "")
	flag.StringVar(&configSecKey, "config.secret.key", "config.yaml", "")
	flag.StringVar(&configFile, "config.file", "config.yaml", "")
	flag.StringVar(&serviceNS, "service.namespace", "infra", "The namespace that the controllers service is registered in")
	flag.StringVar(&serviceName, "service.name", "prom-config-controller", "The controllers service name")
	flag.StringVar(&tlsKey, "tls.key", "tls.key", "Path to TLS key file")
	flag.StringVar(&tlsCert, "tls.cert", "tls.crt", "Path to TLS public cert file")
	flag.StringVar(&tlsCA, "tls.ca", "ca.crt", "Path to TLS CA file")
	flag.StringVar(&reloadScheme, "reload.scheme", "http", "On config change, Reload")
	flag.StringVar(&reloadMethod, "reload.method", "POST", "On config change, Reload")
	flag.StringVar(&reloadHost, "reload.host", "localhost:9090", "host:port for reload, port is used with endpoint mathcing, if host is non-blank, a request will be made direct to the given port.")
	flag.StringVar(&reloadPath, "reload.path", "/-/reload", "On config change, Reload")
	flag.StringVar(&reloadEndpointsNS, "reload.endpointsns", "", "On config change, Reload")
	flag.StringVar(&reloadEndpoints, "reload.endpoints", "", "On config change, Reload")
	flag.DurationVar(&reloadDelay, "reload.delay", 2*time.Second, "delay to allow configmap changes to propagate")
	flag.IntVar(&reloadRetries, "reload.retries", 4, "number of retries when reloading")

	flag.StringVar(&gcpProject, "gcpProject", "", "Google Cloud project to scan for GKE clusters")
	flag.StringVar(&gcpKeysDir, "gcpKeysDir", ".", "directory to write client keys to")
}

type filteredLog struct {
	base io.Writer
}

func (fl *filteredLog) Write(bs []byte) (int, error) {
	if bytes.Contains(bs, []byte("TLS handshake error")) {
		return len(bs), nil
	}
	return fl.base.Write(bs)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	promClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building prometheus clientset: %s", err.Error())
	}

	promInformerFactory := informers.NewFilteredSharedInformerFactory(
		promClient,
		time.Second*30,
		namespace,
		func(opt *metav1.ListOptions) {
			opt.LabelSelector = selector
		},
	)

	sel, err := labels.Parse(selector)
	if err != nil {
		glog.Fatalf("error parsing selector, %v", err)
	}

	tlsCABytes, err := ioutil.ReadFile(tlsCA)
	if err != nil {
		glog.Fatalf("error reading ca cert, %v", err)
	}

	var tmpl *template.Template
	if configTemplate != "" {
		glog.Infof("parsing template %q", configTemplate)
		tmpl, err = template.New(path.Base(configTemplate)).Funcs(sprig.TxtFuncMap()).ParseFiles(configTemplate)
		if err != nil {
			glog.Infof("parsing template failed, %v", err)
			glog.Fatalf("error parsing config template, %v", err)
		}
	}

	host, port, err := net.SplitHostPort(reloadHost)
	if err != nil {
		glog.Fatalf("error parsing host:port pair, %v", err)
	}

	reloader := &reloader{
		method:  reloadMethod,
		path:    reloadPath,
		scheme:  reloadScheme,
		host:    host,
		port:    port,
		delay:   reloadDelay,
		retries: reloadRetries,

		client:    kubeClient,
		name:      reloadEndpoints,
		namespace: reloadEndpointsNS,
	}

	ccfg := ControllerConfig{
		CACert:           tlsCABytes,
		ServiceNS:        serviceNS,
		ServiceName:      serviceName,
		Namespace:        namespace,
		Selector:         sel,
		ConfigTemplate:   tmpl,
		RuleConfigMapNS:  rulesMapNS,
		RuleConfigMap:    rulesMapName,
		RuleConfigMapKey: rulesMapKey,
		RuleFile:         rulesFile,
		ConfigSecretNS:   configSecNS,
		ConfigSecret:     configSecName,
		ConfigSecretKey:  configSecKey,
		ConfigFile:       configFile,
	}

	tokenFile := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	var cl clusterLister
	if gcpProject != "" {
		gkecm, err := gke.NewClusterManagerClient(context.Background())
		if err != nil {
			glog.Fatalf("could not build GKE client, %s", err.Error())
		}

		tokenFile = filepath.Join(gcpKeysDir, "token")
		go func() {
			tokenGrace := 5 * time.Minute
			creds, err := google.FindDefaultCredentials(context.Background(), "https://www.googleapis.com/auth/cloud-platform")
			if err != nil {
				glog.Fatalf("could not get google token source", err.Error())
			}
			for {
				tok, err := creds.TokenSource.Token()
				if err != nil {
					glog.Fatalf("could not get google token source", err.Error())
				}
				ioutil.WriteFile(tokenFile, []byte(tok.AccessToken), 0600)
				if tok.Expiry.IsZero() {
					return
				}
				time.Sleep(time.Until(tok.Expiry) - tokenGrace)
			}
		}()
		cl = listGKEClusters(gkecm, gcpProject, gcpKeysDir, tokenFile)
	}

	controller := NewController(
		ccfg,
		kubeClient,
		promClient,
		promInformerFactory,
		reloader,
		cl,
	)

	go promInformerFactory.Start(stopCh)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "OK") })
	mux.HandleFunc("/validate", serveValidate)

	// Best practice TLS setup: https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	logger := log.New(&filteredLog{os.Stderr}, "", log.LstdFlags)
	s := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
		ErrorLog:     logger,
	}

	go func() {
		<-stopCh
		glog.Info("stopCh closed")
		s.Shutdown(context.Background())
	}()

	g := &errgroup.Group{}
	g.Go(func() error { return s.ListenAndServeTLS(tlsCert, tlsKey) })
	g.Go(func() error { return controller.Run(stopCh) })

	if err := g.Wait(); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
