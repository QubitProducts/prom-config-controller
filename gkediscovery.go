package main

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	container "google.golang.org/api/container/v1"
)

type clusterLister func(ctx context.Context) ([]*container.Cluster, error)

func listGKEClusters(client *http.Client, project string) clusterLister {
	return func(ctx context.Context) ([]*container.Cluster, error) {
		svc, err := container.New(client)
		if err != nil {
			return nil, errors.Wrap(err, "could not create GKE client")
		}
		clusterList, err := svc.Projects.Zones.Clusters.List(project, "-").Context(ctx).Do()
		if err != nil {
			return nil, errors.Wrap(err, "could not list cluster")
		}
		return clusterList.Clusters, nil
	}
}
