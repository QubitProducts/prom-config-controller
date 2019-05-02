package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	gke "cloud.google.com/go/container/apiv1"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	gkepb "google.golang.org/genproto/googleapis/container/v1"
)

type cluster struct {
	Project         string
	CertFile        string
	KeyFile         string
	CAFile          string
	BearerTokenFile string
	gkepb.Cluster
}

type clusterLister func(ctx context.Context) ([]*cluster, error)

func listGKEClusters(c *gke.ClusterManagerClient, project string, tlsDir string, tokenFile string) clusterLister {
	return func(ctx context.Context) ([]*cluster, error) {
		req := &gkepb.ListClustersRequest{
			Parent: "projects/" + project + "/locations/-",
		}

		resp, err := c.ListClusters(ctx, req)

		if err != nil {
			return nil, errors.Wrap(err, "could not list cluster")
		}
		var cls []*cluster
		for _, c := range resp.Clusters {
			newcluster := &cluster{
				Project:         project,
				Cluster:         *c,
				BearerTokenFile: tokenFile,
			}

			if len(c.MasterAuth.ClientCertificate) != 0 {
				crtFile := filepath.Join(tlsDir, c.Name+".crt")
				bs, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClientCertificate)
				if err != nil {
					glog.Errorf("failed decoding %s cert, %v", c.Name, err)
				}
				err = ioutil.WriteFile(crtFile, bs, 0600)
				if err != nil {
					glog.Errorf("failed writing %s, %v", crtFile, err)
				}
				newcluster.CertFile = crtFile
			}

			if len(c.MasterAuth.ClientKey) != 0 {
				keyFile := filepath.Join(tlsDir, c.Name+".key")
				bs, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClientKey)
				if err != nil {
					glog.Errorf("failed decoding %s key, %v", c.Name, err)
				}
				err = ioutil.WriteFile(keyFile, bs, 0600)
				if err != nil {
					glog.Errorf("failed writing %s, %v", keyFile, err)
				}
				newcluster.KeyFile = keyFile
			}

			if len(c.MasterAuth.ClusterCaCertificate) != 0 {
				cacrtFile := filepath.Join(tlsDir, c.Name+".cacrt")
				bs, err := base64.StdEncoding.DecodeString(c.MasterAuth.ClusterCaCertificate)
				if err != nil {
					glog.Errorf("failed decoding %s ca cert, %v", c.Name, err)
				}
				err = ioutil.WriteFile(cacrtFile, bs, 0600)
				if err != nil {
					glog.Errorf("failed writing %s, %v", cacrtFile, err)
				}
				newcluster.CAFile = cacrtFile
			}

			cls = append(cls, newcluster)
		}

		return cls, nil
	}
}
