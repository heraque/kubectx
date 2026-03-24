// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubeconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ahmetb/kubectx/internal/cmdutil"
	"github.com/google/go-cmp/cmp"
)

func Test_kubeconfigPaths_default(t *testing.T) {
	t.Setenv("KUBECONFIG", "")
	t.Setenv("HOME", "/x/y/z")
	t.Setenv("USERPROFILE", "")

	expected := filepath.FromSlash("/x/y/z/.kube/config")
	got, err := kubeconfigPaths()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != expected {
		t.Fatalf("got=%q expected=[%q]", got, expected)
	}
}

func Test_kubeconfigPaths_discoversFilesInKubeDir(t *testing.T) {
	home := t.TempDir()
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(kubeDir, "config")
	otherPath := filepath.Join(kubeDir, "team-b")
	ignoredPath := filepath.Join(kubeDir, "kubectx")

	if err := os.WriteFile(configPath, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(otherPath, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ignoredPath, []byte("previous-context"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("KUBECONFIG", "")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")

	got, err := kubeconfigPaths()
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{configPath, otherPath}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func Test_kubeconfigPaths_noEnvVars(t *testing.T) {
	t.Setenv("KUBECONFIG", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	_, err := kubeconfigPaths()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func Test_kubeconfigPaths_envSingleFile(t *testing.T) {
	home := t.TempDir()
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatal(err)
	}
	discovered := filepath.Join(kubeDir, "extra")
	if err := os.WriteFile(discovered, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	t.Setenv("KUBECONFIG", "foo")

	v, err := kubeconfigPaths()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"foo", discovered}
	if diff := cmp.Diff(expected, v); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func Test_kubeconfigPaths_envMultipleFiles(t *testing.T) {
	home := t.TempDir()
	kubeDir := filepath.Join(home, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		t.Fatal(err)
	}
	discovered := filepath.Join(kubeDir, "extra")
	if err := os.WriteFile(discovered, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}

	path := strings.Join([]string{"file1", "file2", "file3"}, string(os.PathListSeparator))
	t.Setenv("KUBECONFIG", path)
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")

	v, err := kubeconfigPaths()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"file1", "file2", "file3", discovered}
	if diff := cmp.Diff(expected, v); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func TestStandardKubeconfigLoader_returnsNotFoundErr(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KUBECONFIG", "foo")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", "")
	kc := new(Kubeconfig).WithLoader(DefaultLoader)
	err := kc.Parse()
	if err == nil {
		t.Fatal("expected err")
	}
	if !cmdutil.IsNotFoundErr(err) {
		t.Fatalf("expected ENOENT error; got=%v", err)
	}
}

func TestStandardKubeconfigLoader_multipleFiles_skipsMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", "")
	existing := filepath.Join(dir, "config1")
	if err := os.WriteFile(existing, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "config2")

	path := strings.Join([]string{existing, missing}, string(os.PathListSeparator))
	t.Setenv("KUBECONFIG", path)

	files, err := new(StandardKubeconfigLoader).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	files[0].Close()
}

func TestStandardKubeconfigLoader_multipleFiles_allMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", "")
	path := strings.Join([]string{
		filepath.Join(dir, "missing1"),
		filepath.Join(dir, "missing2"),
	}, string(os.PathListSeparator))
	t.Setenv("KUBECONFIG", path)

	_, err := new(StandardKubeconfigLoader).Load()
	if err == nil {
		t.Fatal("expected error when all files missing")
	}
	if !cmdutil.IsNotFoundErr(err) {
		t.Fatalf("expected ENOENT error; got=%v", err)
	}
}

func TestStandardKubeconfigLoader_multipleFiles_loadsAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", "")
	f1 := filepath.Join(dir, "config1")
	f2 := filepath.Join(dir, "config2")
	if err := os.WriteFile(f1, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("apiVersion: v1\nkind: Config\ncontexts: []\n"), 0644); err != nil {
		t.Fatal(err)
	}

	path := strings.Join([]string{f1, f2}, string(os.PathListSeparator))
	t.Setenv("KUBECONFIG", path)

	files, err := new(StandardKubeconfigLoader).Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	for _, f := range files {
		f.Close()
	}
}
