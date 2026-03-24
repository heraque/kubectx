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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/ahmetb/kubectx/internal/cmdutil"
	"gopkg.in/yaml.v3"
)

var (
	DefaultLoader Loader = new(StandardKubeconfigLoader)
)

type StandardKubeconfigLoader struct{}

type kubeconfigFile struct {
	*os.File
	path string
}

func (kf *kubeconfigFile) Path() string { return kf.path }

func (*StandardKubeconfigLoader) Load() ([]ReadWriteResetCloser, error) {
	paths, err := kubeconfigPaths()
	if err != nil {
		return nil, fmt.Errorf("cannot determine kubeconfig path: %w", err)
	}

	var files []ReadWriteResetCloser
	for _, p := range paths {
		f, err := os.OpenFile(p, os.O_RDWR, 0)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to open file %q: %w", p, err)
		}
		files = append(files, &kubeconfigFile{File: f, path: p})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("kubeconfig file not found: %w",
			&os.PathError{Op: "open", Path: paths[0], Err: os.ErrNotExist})
	}
	return files, nil
}

func (kf *kubeconfigFile) Reset() error {
	if err := kf.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	if _, err := kf.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek in file: %w", err)
	}
	return nil
}

func kubeconfigPaths() ([]string, error) {
	// KUBECONFIG env var
	if v := os.Getenv("KUBECONFIG"); v != "" {
		return filepath.SplitList(v), nil
	}

	home := cmdutil.HomeDir()
	if home == "" {
		return nil, errors.New("HOME or USERPROFILE environment variable not set")
	}

	kubeDir := filepath.Join(home, ".kube")
	paths, err := discoverKubeconfigPaths(kubeDir)
	if err != nil {
		return nil, err
	}
	if len(paths) > 0 {
		return paths, nil
	}
	return []string{filepath.Join(kubeDir, "config")}, nil
}

func discoverKubeconfigPaths(kubeDir string) ([]string, error) {
	entries, err := os.ReadDir(kubeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read kubeconfig directory %q: %w", kubeDir, err)
	}

	configPath := filepath.Join(kubeDir, "config")
	var extraPaths []string

	for _, entry := range entries {
		if entry.Name() == "config" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to stat %q: %w", filepath.Join(kubeDir, entry.Name()), err)
		}
		if !info.Mode().IsRegular() {
			continue
		}

		path := filepath.Join(kubeDir, entry.Name())
		ok, err := isKubeconfigFile(path)
		if err != nil {
			return nil, err
		}
		if ok {
			extraPaths = append(extraPaths, path)
		}
	}

	slices.Sort(extraPaths)

	info, err := os.Stat(configPath)
	switch {
	case err == nil && info.Mode().IsRegular():
		return append([]string{configPath}, extraPaths...), nil
	case err == nil:
		return extraPaths, nil
	case os.IsNotExist(err):
		return extraPaths, nil
	default:
		return nil, fmt.Errorf("failed to stat %q: %w", configPath, err)
	}
}

func isKubeconfigFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open %q while scanning kubeconfigs: %w", path, err)
	}
	defer f.Close()

	var node yaml.Node
	if err := yaml.NewDecoder(f).Decode(&node); err != nil {
		return false, nil
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return false, nil
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return false, nil
	}

	for i := 0; i+1 < len(root.Content); i += 2 {
		switch root.Content[i].Value {
		case "apiVersion", "kind", "clusters", "contexts", "users", "current-context":
			return true, nil
		}
	}
	return false, nil
}
