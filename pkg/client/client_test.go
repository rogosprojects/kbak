package client

import (
	"os"
	"path/filepath"
	"testing"
)

const testKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: lab
  cluster: {server: "https://lab.example.com:6443"}
users:
- name: admin
  user: {token: fake}
contexts:
- name: lab
  context: {cluster: lab, user: admin, namespace: team-a}
current-context: lab
`

func TestNewClientUsesKubeconfigEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(testKubeconfig), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", path)
	t.Setenv("HOME", t.TempDir()) // no ~/.kube/config to fall back on

	c, err := NewClient("", false)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if c.Config.Host != "https://lab.example.com:6443" {
		t.Errorf("Host = %q, want the server from $KUBECONFIG", c.Config.Host)
	}
	if c.Namespace != "team-a" {
		t.Errorf("Namespace = %q, want %q", c.Namespace, "team-a")
	}
}

func TestNewClientMissingExplicitKubeconfig(t *testing.T) {
	if _, err := NewClient(filepath.Join(t.TempDir(), "missing"), false); err == nil {
		t.Error("NewClient() with a missing --kubeconfig file should fail")
	}
}
