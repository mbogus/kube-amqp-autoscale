// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import (
	"errors"
	"testing"
)

func TestNewSizeNoCoverage(t *testing.T) {
	sc := scaleContext{Coverage: 0.75}
	_, err := sc.newSize(0.0, 0.5)
	if err == nil {
		t.Fatal("Expected error")
	}
	if got, want := err.Error(), "not enough metrics to calculate new size, required at least 0.75 was 0.50 metrics ratio"; got != want {
		t.Errorf("Expected %s, got: %s", want, got)
	}
}

func TestNewSize(t *testing.T) {
	sc := scaleContext{Coverage: 0.75, Threshold: 1}
	s, err := sc.newSize(2.5, 0.75)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := s, int32(3); got != want {
		t.Errorf("Expected %d, got: %d", want, got)
	}
}

func TestAutoscaleClosedChannel(t *testing.T) {
	forever := make(chan struct{})
	close(forever)
	var err error
	autoscale(func() (*queueMetrics, error) { return nil, errors.New("error") }, &scaleContext{}, forever)
	if err != nil {
		t.Fatal(err)
	}
}
