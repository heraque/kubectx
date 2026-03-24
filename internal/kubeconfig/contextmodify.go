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

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func (k *Kubeconfig) DeleteContextEntry(deleteName string) error {
	_, fileIdx, err := k.contextNodeWithFileIndex(deleteName)
	if err != nil {
		return err
	}

	contexts, err := contextsNodeOf(&k.files[fileIdx])
	if err != nil {
		return err
	}
	return contexts.PipeE(
		yaml.ElementSetter{
			Keys:   []string{"name"},
			Values: []string{deleteName},
		},
	)
}

// ModifyCurrentContext writes current-context to the file that owns the target
// context and clears it from the remaining files to keep multi-file configs
// unambiguous.
func (k *Kubeconfig) ModifyCurrentContext(name string) error {
	if len(k.files) == 0 {
		return errNoFiles
	}

	_, fileIdx, err := k.contextNodeWithFileIndex(name)
	if err != nil {
		if len(k.files) == 1 {
			return k.files[0].config.PipeE(yaml.SetField("current-context", yaml.NewScalarRNode(name)))
		}
		return err
	}

	for i := range k.files {
		if i == fileIdx {
			if err := k.files[i].config.PipeE(yaml.SetField("current-context", yaml.NewScalarRNode(name))); err != nil {
				return err
			}
			continue
		}
		if err := k.files[i].config.PipeE(yaml.SetField("current-context", yaml.NewStringRNode(""))); err != nil {
			return err
		}
	}
	return nil
}

func (k *Kubeconfig) ModifyContextName(old, new string) error {
	context, _, err := k.contextNodeWithFileIndex(old)
	if err != nil {
		return err
	}
	if context == nil {
		return errors.New("\"contexts\" entry is nil")
	}
	return context.PipeE(yaml.SetField("name", yaml.NewScalarRNode(new)))
}
