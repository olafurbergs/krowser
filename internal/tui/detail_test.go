package tui

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/olafurb/krowser/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDecodeSecretValue(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{base64.StdEncoding.EncodeToString([]byte("hello")), "hello"},
		{base64.StdEncoding.EncodeToString([]byte("")), "(empty)"},
		{"not-base64!", "<invalid base64>"},
		{base64.StdEncoding.EncodeToString([]byte{0x00, 0x01, 0xff}), "binary data · 3 bytes"},
		{base64.StdEncoding.EncodeToString([]byte("-----BEGIN CERT-----\nline\n")), "-----BEGIN CERT-----\nline\n"},
	}
	for _, c := range cases {
		if got := decodeSecretValue(c.in); got != c.want {
			t.Errorf("decodeSecretValue(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseSecretData(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"data": map[string]any{
			"token":    base64.StdEncoding.EncodeToString([]byte("abc")),
			"password": base64.StdEncoding.EncodeToString([]byte("hunter2")),
		},
	}}
	d := &detailScreen{res: &k8s.Resource{Plural: "secrets"}, kind: "yaml", obj: obj}
	d.parseSecretData()
	if !d.secretMode {
		t.Fatal("expected secret mode enabled")
	}
	if len(d.secretKeys) != 2 || d.secretKeys[0] != "password" || d.secretKeys[1] != "token" {
		t.Errorf("secretKeys = %v, want sorted [password token]", d.secretKeys)
	}
	if d.secretVals[1] != "abc" {
		t.Errorf("decoded token = %q, want abc", d.secretVals[1])
	}

	d2 := &detailScreen{res: &k8s.Resource{Plural: "pods"}, kind: "yaml", obj: obj}
	d2.parseSecretData()
	if d2.secretMode {
		t.Error("secret mode should not be enabled for non-secret resources")
	}

	d3 := &detailScreen{res: &k8s.Resource{Plural: "secrets"}, kind: "describe", obj: obj}
	d3.parseSecretData()
	if d3.secretMode {
		t.Error("secret mode should not be enabled in describe view")
	}
}

func TestSecretDataView(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"data": map[string]any{
			"password": base64.StdEncoding.EncodeToString([]byte("hunter2")),
		},
	}}
	d := &detailScreen{res: &k8s.Resource{Plural: "secrets", Title: "Secrets"}, kind: "yaml", obj: obj}
	d.width = 80
	d.height = 20
	d.parseSecretData()
	out := d.secretDataView()
	if !strings.Contains(out, "data (1)") || !strings.Contains(out, "password") || !strings.Contains(out, "hunter2") {
		t.Errorf("secretDataView missing expected content: %q", out)
	}
}
