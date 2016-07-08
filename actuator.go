// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"fmt"
	"log"
	"time"
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
			qStats, err := fstats()
			if err != nil {
				log.Println(err)
				continue
			}

			replicas, err := ctx.newSize(qStats.Average, qStats.Coverage)
			if err != nil {
				log.Println(err)
				continue
			}

			err = ctx.Scaler(replicas)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (ctx *scaleContext) newSize(avg float64, cov float64) (int32, error) {
	var replicas int32
	var err error
	if cov < ctx.Coverage {
		err = fmt.Errorf("Not enough metrics to calculate new size, required at least %.2f was %.2f metrics ratio", ctx.Coverage, cov)
	} else {
		replicas = int32(avg / float64(ctx.Threshold))
	}
	return replicas, err
}
