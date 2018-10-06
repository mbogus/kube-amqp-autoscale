// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"fmt"
	"log"
	"time"
	"math"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	desiredReplicas = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "desired_replicas",
		Help:      "Desired number of replicas.",
	})
	autoscaleErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "autoscale_failures_total",
		Help:      "Number of failed autoscale operations.",
	})
	pollCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "polls_total",
		Help:      "Count of times autoscale polling loop runs.",
	})
)

type queueStats func() (*queueMetrics, error)
type scaler func(int32) error

type scaleContext struct {
	Threshold int
	Coverage  float64
	Interval  int

	Scaler scaler
}

func autoscale(
	fstats queueStats,
	ctx *scaleContext,
	quit <-chan struct{}) {
	for {
		select {
		case <-quit:
			return
		case <-time.After(time.Duration(ctx.Interval) * time.Second):
			pollCount.Inc()
			qStats, err := fstats()
			if err != nil {
				log.Println(err)
				autoscaleErrors.Inc()
				continue
			}

			replicas, err := ctx.newSize(qStats.Average, qStats.Coverage)
			if err != nil {
				log.Println(err)
				continue
			}
			desiredReplicas.Set(float64(replicas))

			err = ctx.Scaler(replicas)
			if err != nil {
				log.Println(err)
				autoscaleErrors.Inc()
			}
		}
	}
}

func (ctx *scaleContext) newSize(avg float64, cov float64) (int32, error) {
	var replicas int32
	var err error
	if cov < ctx.Coverage {
		err = fmt.Errorf("not enough metrics to calculate new size, required at least %.2f was %.2f metrics ratio", ctx.Coverage, cov)
	} else {
		replicas = int32(math.Ceil(avg / float64(ctx.Threshold)))
	}
	return replicas, err
}
