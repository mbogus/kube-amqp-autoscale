// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import "testing"

func TestSetVersion(t *testing.T) {
	setVersion()
	if got, want := appVersion, "(untracked build)"; got != want {
		t.Errorf("Expected appVersion='%s', got: '%s'", want, got)
	}

	Version = "0.1"
	BuildType = "RELEASE"
	Build = "deadbeef"
	setVersion()
	if got, want := appVersion, "0.1 (deadbeef)"; got != want {
		t.Errorf("Expected appVersion='%s', got: '%s'", want, got)
	}

	BuildType = "SNAPSHOT"
	setVersion()
	if got, want := appVersion, "0.1-SNAPSHOT (deadbeef)"; got != want {
		t.Errorf("Expected appVersion='%s', got: '%s'", want, got)
	}

	BuildType = "DEV"
	BuildDate = "2016-04-15T10:33:16+0100"
	setVersion()
	if got, want := appVersion, "0.1-DEV (+deadbeef 2016-04-15T10:33:16+0100)"; got != want {
		t.Errorf("Expected appVersion='%s', got: '%s'", want, got)
	}

}

func TestValidateParams(t *testing.T) {
	err := validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing RabbitMQ URI"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	brokerURIParam = "amqp://"

	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing RabbitMQ queue name"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	queueNameParam = "queue"

	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing Kubernetes API URL"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	apiURLParam = "http://"

	intervalParam = 0
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Invalid auto-scale interval '0'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	intervalParam = 10

	thresholdParam = 0
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Invalid threshold value '0'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	thresholdParam = 1

	statsIntervalParam = 15
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Interval for saving statistics '15' should be smaller than auto-scale interval '10'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	statsIntervalParam = 5

	statsCoverageParam = 1.5
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Invalid metrics coverage ratio '1.50'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	statsCoverageParam = 1.0

	minParam = -1
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Invalid lower limit for the number of pods '-1'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}

	minParam = 1
	maxParam = 1
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Upper limit for the number of pods '1' must be greater than lower limit '1'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	maxParam = 2

	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing name of the resource to autoscale"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	nameParam = "pod"

	kindParam = ""
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing kind of the resource to autoscale"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}

	kindParam = "X"
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Invalid kind of the resource 'X'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}

	kindParam = "ReplicationController"
	namespaceParam = ""
	err = validateParams()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "Missing namespace of the resource to autoscale"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}

	namespaceParam = "default"
	err = validateParams()
	if err != nil {
		t.Fatal(err)
	}

}
