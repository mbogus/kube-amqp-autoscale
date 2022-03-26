// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
)

const (
	deploymentKind            = "Deployment"
	replicationControllerKind = "ReplicationController"
	replicaSetKind            = "ReplicaSet"
)

var (
	scalingEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scaling_events_total",
			Help:      "Count of scaling events by resource kind.",
		},
		[]string{"kind", "name"},
	)
)

type apiContext struct {
	URL       string
	User      string
	Passwd    string
	TokenFile string
	CAFile    string
	Insecure  bool
	Bounds    *scaleBounds

	clientConf *restclient.Config
}

func scale(kind string, ns string, name string, newSize int32, ctx *apiContext) error {
	c, err := ctx.client()
	if err != nil {
		return err
	}
	return scaleKind(c, kind, ns, name, newSize, ctx.Bounds)
}

func scaleKind(c *kubernetes.Clientset, kind string, ns string, name string, newSize int32, b *scaleBounds) error {
	switch kind {
	case replicaSetKind:
		return scaleReplicaSets(c, ns, name, newSize, b)
	case deploymentKind:
		return scaleDeployments(c, ns, name, newSize, b)
	}
	return fmt.Errorf("No scaler has been implemented for '%s'", kind)
}

func scaleDeployments(c *kubernetes.Clientset, ns string, name string, newSize int32, b *scaleBounds) error {
	deployment, err := c.AppsV1().Deployments(ns).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}
	replicas := b.newSize(*deployment.Spec.Replicas, newSize)
	if replicas != *deployment.Spec.Replicas {
		log.Printf("Scaling deployment '%s' from %d to %d replicas", name, *deployment.Spec.Replicas, replicas)
		scalingEvents.With(prometheus.Labels{"kind": "Deployment", "name": name}).Inc()
		deployment.Spec.Replicas = &replicas
		_, err = c.AppsV1().Deployments(ns).Update(deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func scaleReplicaSets(c *kubernetes.Clientset, ns string, name string, newSize int32, b *scaleBounds) error {
	pod, err := c.AppsV1().ReplicaSets(ns).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}
	replicas := b.newSize(*pod.Spec.Replicas, newSize)
	if replicas != *pod.Spec.Replicas {
		log.Printf("Scaling replica set '%s' from %d to %d replicas", name, *pod.Spec.Replicas, replicas)
		scalingEvents.With(prometheus.Labels{"kind": "ReplicaSet", "name": name}).Inc()
		pod.Spec.Replicas = &replicas
		_, err = c.AppsV1().ReplicaSets(ns).Update(pod)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *apiContext) client() (*kubernetes.Clientset, error) {
	if ctx.clientConf == nil {
		conf, err := apiConfig(ctx.URL, ctx.User, ctx.Passwd, ctx.TokenFile, ctx.CAFile, ctx.Insecure)
		if err != nil {
			return nil, err
		}
		ctx.clientConf = conf
	}
	return kubernetes.NewForConfig(ctx.clientConf)
}

func apiConfig(apiURL string,
	apiUser string,
	apiPasswd string,
	apiTokenFile string,
	apiCAFile string,
	apiInsecure bool) (*restclient.Config, error) {

	if len(apiURL) == 0 {
		return nil, fmt.Errorf("API URL must be defined")
	}

	cfg := &restclient.Config{
		Host: apiURL,
	}

	if apiInsecure {
		cfg.Insecure = apiInsecure
	}

	if len(apiUser) > 0 && len(apiPasswd) > 0 {
		cfg.Username = apiUser
		cfg.Password = apiPasswd
	}

	if len(apiTokenFile) > 0 {
		token, err := ioutil.ReadFile(apiTokenFile)
		if err != nil {
			return nil, err
		}
		cfg.BearerToken = string(token)
	}

	if len(apiCAFile) > 0 {
		tlsClientConfig := restclient.TLSClientConfig{}
		if _, err := certutil.NewPool(apiCAFile); err != nil {
			log.Printf("Expected to load root CA config from '%s', but got err: %v", apiCAFile, err)
		} else {
			tlsClientConfig.CAFile = apiCAFile
			cfg.TLSClientConfig = tlsClientConfig
		}
	}

	return cfg, nil
}

type scaleBounds struct {
	Min           int
	Max           int
	IncreaseLimit int
	DecreaseLimit int
}

func (sb *scaleBounds) newSize(size, newSize int32) int32 {
	if size == newSize {
		return newSize
	}
	replicas := newSize
	if newSize < size {
		if sb.DecreaseLimit > 0 {
			replicas = max(size-int32(sb.DecreaseLimit), newSize)
		}
	} else {
		if sb.IncreaseLimit > 0 {
			replicas = min(size+int32(sb.IncreaseLimit), newSize)
		}
	}
	replicas = max(replicas, int32(sb.Min))
	replicas = min(replicas, int32(sb.Max))

	return replicas
}

func min(x, y int32) int32 {
	if x < y {
		return x
	}
	return y
}

func max(x, y int32) int32 {
	if x < y {
		return y
	}
	return x
}
