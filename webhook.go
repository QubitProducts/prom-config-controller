package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	configV1beta1 "github.com/QubitProducts/prom-config-controller/pkg/apis/config/v1beta1"
	"github.com/golang/glog"
	v1 "k8s.io/api/admission/v1"
	regv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var codecs = serializer.NewCodecFactory(scheme.Scheme)

type admitFunc func(v1.AdmissionReview) *v1.AdmissionResponse

func toAdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func serveValidate(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Info("Webhook called")
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	var reviewResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Error(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = admit(ar)
	}

	response := v1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	resp, err := json.Marshal(response)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func admit(ar v1.AdmissionReview) *v1.AdmissionResponse {
	glog.V(2).Info("admitting prometheus resource")
	if ar.Request.Resource.Group != configV1beta1.SchemeGroupVersion.Group ||
		ar.Request.Resource.Version != configV1beta1.SchemeGroupVersion.Version {
		err := errors.New("unexpected resource or version")
		glog.Error(err)
		return toAdmissionResponse(err)
	}

	switch ar.Request.Resource.Resource {
	case "rulegroups":
		return admitRuleGroups(ar)
	case "scrapes":
		return admitScrapes(ar)
	default:
		err := fmt.Errorf("unknown resource %s", ar.Request.Resource.Resource)
		glog.Error(err)
		return toAdmissionResponse(err)
	}
}

func admitRuleGroups(ar v1.AdmissionReview) *v1.AdmissionResponse {
	glog.V(2).Info("admitting prometheus rule group")

	raw := ar.Request.Object.Raw
	rulegroup := configV1beta1.RuleGroup{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &rulegroup); err != nil {
		glog.Error(err)
		return toAdmissionResponse(err)
	}
	reviewResponse := v1.AdmissionResponse{
		Allowed: true,
	}

	_, errs := convertRuleGroup(rulegroup.GetName(), &rulegroup)
	if len(errs) == 0 {
		return &reviewResponse
	}
	var messages []string
	causes := make([]metav1.StatusCause, len(errs))
	for _, e := range errs {
		causes = append(causes, metav1.StatusCause{
			Message: e.Error(),
		})
		messages = append(messages, e.Error())
	}

	reviewResponse.Allowed = false
	reviewResponse.Result = &metav1.Status{
		Message: fmt.Sprintf("errors during rulegroup validation, %s", strings.Join(messages, ", ")),
		Reason:  metav1.StatusReasonNotAcceptable,
		Details: &metav1.StatusDetails{
			Causes: causes,
		},
	}
	return &reviewResponse
}

func admitScrapes(ar v1.AdmissionReview) *v1.AdmissionResponse {
	glog.V(2).Info("admitting prometheus scrape")

	raw := ar.Request.Object.Raw
	scrape := configV1beta1.Scrape{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &scrape); err != nil {
		glog.Error(err)
		return toAdmissionResponse(err)
	}
	reviewResponse := v1.AdmissionResponse{
		Allowed: true,
	}

	_, err := convertScrape(scrape.GetName(), &scrape)
	if err == nil {
		return &reviewResponse
	}
	causes := []metav1.StatusCause{{
		Message: err.Error(),
	},
	}

	reviewResponse.Allowed = false
	reviewResponse.Result = &metav1.Status{
		Message: fmt.Sprintf("errors during scrape validation, %s", err.Error()),
		Reason:  metav1.StatusReasonNotAcceptable,
		Details: &metav1.StatusDetails{
			Causes: causes,
		},
	}
	return &reviewResponse
}

// register this webhook admission controller with the kube-apiserver
// by creating MutatingWebhookConfiguration.
func (c *Controller) selfRegistration() error {
	ctx := context.Background()
	webhookName := "prom-config-controller"
	path := "/validate"
	time.Sleep(10 * time.Second)
	client := c.kubeclientset.AdmissionregistrationV1().ValidatingWebhookConfigurations()
	_, err := client.Get(ctx, webhookName, metav1.GetOptions{})
	if err == nil {
		if err2 := client.Delete(ctx, webhookName, metav1.DeleteOptions{}); err2 != nil {
			return err2
		}
	}
	webhookConfig := &regv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []regv1.ValidatingWebhook{
			{
				Name: configV1beta1.SchemeGroupVersion.Group,
				Rules: []regv1.RuleWithOperations{
					{
						Operations: []regv1.OperationType{regv1.Create, regv1.Update},
						Rule: regv1.Rule{
							APIGroups:   []string{configV1beta1.SchemeGroupVersion.Group},
							APIVersions: []string{configV1beta1.SchemeGroupVersion.Version},
							Resources:   []string{"rulegroups", "scrapes"},
						},
					}},
				ClientConfig: regv1.WebhookClientConfig{
					Service: &regv1.ServiceReference{
						Namespace: c.ServiceNS,
						Name:      c.ServiceName,
						Path:      &path,
					},
					CABundle: c.CACert,
				},
			},
		},
	}
	if _, err := client.Create(ctx, webhookConfig, metav1.CreateOptions{}); err != nil {
		return err
	}

	glog.V(2).Info("Self registration as MutatingWebhook succeeded.")
	return nil
}
