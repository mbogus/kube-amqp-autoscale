// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/crypto"
)

const (
	deploymentKind            = "Deployment"
	jobKind                   = "Job"
	replicationControllerKind = "ReplicationController"
	replicaSetKind            = "ReplicaSet"
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

func scaleKind(c *client.Client, kind string, ns string, name string, newSize int32, b *scaleBounds) error {
	switch kind {
	case replicationControllerKind:
		return scaleReplicationControllers(c, ns, name, newSize, b)
	case replicaSetKind:
		return scaleReplicaSets(c, ns, name, newSize, b)
	case deploymentKind:
		return scaleDeployments(c, ns, name, newSize, b)
	}
	return fmt.Errorf("No scaler has been implemented for '%s'", kind)
}

func scaleDeployments(c *client.Client, ns string, name string, newSize int32, b *scaleBounds) error {
	pod, err := c.Deployments(ns).Get(name)
	if err != nil {
		return err
	}
	replicas := b.newSize(pod.Spec.Replicas, newSize)
	if replicas != pod.Spec.Replicas {
		log.Printf("Scaling deployment '%s' from %d to %d replicas", name, pod.Spec.Replicas, replicas)
		pod.Spec.Replicas = replicas
		_, err = c.Deployments(ns).Update(pod)
		if err != nil {
			return err
		}
	}
	return nil
}

func scaleReplicationControllers(c *client.Client, ns string, name string, newSize int32, b *scaleBounds) error {
	pod, err := c.ReplicationControllers(ns).Get(name)
	if err != nil {
		return err
	}
	replicas := b.newSize(pod.Spec.Replicas, newSize)
	if replicas != pod.Spec.Replicas {
		log.Printf("Scaling replication controller '%s' from %d to %d replicas", name, pod.Spec.Replicas, replicas)
		pod.Spec.Replicas = replicas
		_, err = c.ReplicationControllers(ns).Update(pod)
		if err != nil {
			return err
		}
	}
	return nil
}

func scaleReplicaSets(c *client.Client, ns string, name string, newSize int32, b *scaleBounds) error {
	pod, err := c.ReplicaSets(ns).Get(name)
	if err != nil {
		return err
	}
	replicas := b.newSize(pod.Spec.Replicas, newSize)
	if replicas != pod.Spec.Replicas {
		log.Printf("Scaling replica set '%s' from %d to %d replicas", name, pod.Spec.Replicas, replicas)
		pod.Spec.Replicas = replicas
		_, err = c.ReplicaSets(ns).Update(pod)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *apiContext) client() (*client.Client, error) {
	if ctx.clientConf == nil {
		conf, err := apiConfig(ctx.URL, ctx.User, ctx.Passwd, ctx.TokenFile, ctx.CAFile, ctx.Insecure)
		if err != nil {
			return nil, err
		}
		ctx.clientConf = conf
	}
	return client.New(ctx.clientConf)
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
		if _, err := crypto.CertPoolFromFile(apiCAFile); err != nil {
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
