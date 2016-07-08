// Copyright (c) 2016, M Bogus.
// This source file is part of the KUBE-AMQP-AUTOSCALE open source project
// Licensed under Apache License v2.0
// See LICENSE file for license information

package main

import "testing"

func TestScaleInvalidKind(t *testing.T) {
	if got, want := scale("X", "", "", 0, &apiContext{URL: "http://127.0.0.1:8080"}).Error(), "No scaler has been implemented for 'X'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestScaleKindInvalidKind(t *testing.T) {
	if got, want := scaleKind(nil, "X", "", "", 0, nil).Error(), "No scaler has been implemented for 'X'"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestAPIConfigNoURL(t *testing.T) {
	_, err := apiConfig("", "", "", "", "", false)
	if got, want := err.Error(), "API URL must be defined"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestAPIConfigInsecure(t *testing.T) {
	c, err := apiConfig("https://127.0.0.1:443", "", "", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := c.Host, "https://127.0.0.1:443"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	if got, want := c.Insecure, true; got != want {
		t.Errorf("Expected error='%v', got: '%v'", want, got)
	}
}

func TestAPIConfigBasicAuth(t *testing.T) {
	c, err := apiConfig("http://127.0.0.1:8080", "KubeUser", "KubePasswd", "", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := c.Host, "http://127.0.0.1:8080"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	if got, want := c.Insecure, false; got != want {
		t.Errorf("Expected error='%v', got: '%v'", want, got)
	}
	if got, want := c.Username, "KubeUser"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
	if got, want := c.Password, "KubePasswd"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestAPIConfigTokenNonExistentFile(t *testing.T) {
	_, err := apiConfig("http://127.0.0.1:8080", "", "", "/tmp/file-that-does-not-exist", "", false)
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "open /tmp/file-that-does-not-exist: no such file or directory"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestAPIContextClientNoURL(t *testing.T) {
	ctx := apiContext{}
	_, err := ctx.client()
	if err == nil {
		t.Fatalf("Expected error")
	}
	if got, want := err.Error(), "API URL must be defined"; got != want {
		t.Errorf("Expected error='%s', got: '%s'", want, got)
	}
}

func TestAPIContextClientURL(t *testing.T) {
	ctx := apiContext{URL: "http://127.0.0.1:8080"}
	_, err := ctx.client()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMaxEqual(t *testing.T) {
	if got, want := max(0, 0), int32(0); got != want {
		t.Errorf("Expected max='%d', got: '%d'", want, got)
	}
}

func TestMaxFirstGreater(t *testing.T) {
	if got, want := max(2, 1), int32(2); got != want {
		t.Errorf("Expected max='%d', got: '%d'", want, got)
	}
}

func TestMaxSecondGreater(t *testing.T) {
	if got, want := max(3, 4), int32(4); got != want {
		t.Errorf("Expected max='%d', got: '%d'", want, got)
	}
}

func TestMinEqual(t *testing.T) {
	if got, want := min(0, 0), int32(0); got != want {
		t.Errorf("Expected min='%d', got: '%d'", want, got)
	}
}

func TestMinFirstGreater(t *testing.T) {
	if got, want := min(2, 1), int32(1); got != want {
		t.Errorf("Expected min='%d', got: '%d'", want, got)
	}
}

func TestMinSecondGreater(t *testing.T) {
	if got, want := min(3, 4), int32(3); got != want {
		t.Errorf("Expected min='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeEqual(t *testing.T) {
	b := scaleBounds{}
	if got, want := b.newSize(int32(1), int32(1)), int32(1); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeGreater(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10}
	if got, want := b.newSize(int32(1), int32(2)), int32(2); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeGreaterThanMax(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10}
	if got, want := b.newSize(int32(1), int32(11)), int32(10); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeGreaterExceedLimit(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, IncreaseLimit: 3}
	if got, want := b.newSize(int32(1), int32(6)), int32(4); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeGreaterWithinLimit(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, IncreaseLimit: 5}
	if got, want := b.newSize(int32(1), int32(3)), int32(3); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeGreaterWithLimitExceedMax(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, IncreaseLimit: 4}
	if got, want := b.newSize(int32(9), int32(12)), int32(10); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeSmaller(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10}
	if got, want := b.newSize(int32(2), int32(1)), int32(1); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeSmallerWithinLimit(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, DecreaseLimit: 3}
	if got, want := b.newSize(int32(6), int32(4)), int32(4); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeSmallerExceedLimit(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, DecreaseLimit: 3}
	if got, want := b.newSize(int32(9), int32(1)), int32(6); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeSmallerThanMinWithLimit(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10, DecreaseLimit: 3}
	if got, want := b.newSize(int32(2), int32(-1)), int32(0); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}

func TestScaleBoundsNewSizeSmallerThanMin(t *testing.T) {
	b := scaleBounds{Min: 0, Max: 10}
	if got, want := b.newSize(int32(2), int32(-1)), int32(0); got != want {
		t.Errorf("Expected newSize='%d', got: '%d'", want, got)
	}
}
