package main

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudflare/backoff"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// A Reloader can reload a promethues deployment on demand.
type Reloader interface {
	Reload()
}

type reloader struct {
	scheme  string
	method  string
	path    string
	host    string
	port    string
	retries int
	delay   time.Duration

	client    kubernetes.Interface
	name      string
	namespace string
}

func (r *reloader) Reload() {
	glog.Infof("reloading prometheus")

	time.Sleep(r.delay)

	var actioned bool
	if r.host != "" {
		r.reloadURL()
		actioned = true
	}

	if r.name != "" && r.namespace != "" {
		r.reloadEndpoints()
		actioned = true
	}

	if !actioned {
		glog.V(2).Info("no reload options are active")
	}
}

func (r *reloader) reloadURL() {
	glog.V(2).Info("performing single reload")
	u := &url.URL{
		Scheme: r.scheme,
		Host:   net.JoinHostPort(r.host, r.port),
		Path:   r.path,
	}

	go reloadOneURL(r.method, u, r.retries)
}

func (r *reloader) reloadEndpoints() {
	ctx := context.Background()
	glog.V(2).Info("performing endpoint reload")
	eps, err := r.client.CoreV1().Endpoints(r.namespace).Get(ctx, r.name, metav1.GetOptions{})
	if err != nil {
		glog.V(1).Infof("failed listing reload endpoints, %v", err)
		return
	}
	if len(eps.Subsets) == 0 {
		glog.V(2).Infof("no endpoints returned for %s/%s", r.namespace, r.name)
		return
	}
	for _, ep := range eps.Subsets {
		for _, addr := range ep.Addresses {
			u := &url.URL{
				Scheme: r.scheme,
				Host:   net.JoinHostPort(addr.IP, r.port),
				Path:   r.path,
			}

			go reloadOneURL(r.method, u, r.retries)
		}
	}
}

func reloadOneURL(method string, u *url.URL, retries int) {
	glog.V(1).Infof("starting reload of %s using a %s", u.String(), method)
	b := backoff.New(0, 0)
	defer b.Reset()
	for retries >= 0 {
		req, err := http.NewRequest(method, u.String(), nil)
		if err != nil {
			glog.V(1).Infof("reload failed, %v", err)
			return
		}

		res, err := http.DefaultClient.Do(req)
		if err == nil {
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()

			if res.StatusCode < 500 {
				glog.V(2).Infof("reloaded %s", u)
				return
			}
			glog.V(1).Infof("reload %s response, %v", u, res.StatusCode)
		}

		glog.V(1).Infof("reload failed, %v", err)

		retries--
		<-time.After(b.Duration())
	}
	glog.V(1).Infof("giving up reload of %s", u)
}
