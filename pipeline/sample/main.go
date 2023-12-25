// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/nautes-labs/nautes/app/runtime-operator/pkg/component"
	"github.com/nautes-labs/nautes/app/runtime-operator/pkg/pipeline/shared"
	"github.com/nautes-labs/nautes/pkg/resource"
	"github.com/nautes-labs/nautes/pkg/thirdpartapis/tekton/pipeline/v1alpha1"
	"github.com/nautes-labs/nautes/pkg/thirdpartapis/tekton/pipeline/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

const HookName = "ls"
const supportPipelineType = "tekton"

var supportEventSourceTypes = []string{"gitlab"}
var (
	varNamePrintPath = "printPath"
	varNameImageName = "imageName"
)

var appLogger = hclog.New(&hclog.LoggerOptions{
	Name:  HookName,
	Level: hclog.LevelFromString("DEBUG"),
})

type report struct{}

func (report) GetPipelineType() (string, error) {
	return supportPipelineType, nil
}

func (report) GetHooksMetadata() ([]resource.HookMetadata, error) {
	metadata := resource.HookMetadata{
		Name:                    HookName,
		IsPreHook:               true,
		IsPostHook:              true,
		SupportEventSourceTypes: supportEventSourceTypes,
		VarsDefinition: &apiextensions.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensions.JSONSchemaProps{
				varNamePrintPath: {
					Type:      "string",
					MaxLength: newInt64(20),
				},
				varNameImageName: {
					Type: "string",
				},
			},
		},
	}

	return []resource.HookMetadata{metadata}, nil
}

func newInt64(num int) *int64 {
	tmp := int64(num)
	return &tmp
}

func (report) BuildHook(hookName string, info component.HookBuildData) (*component.Hook, error) {
	switch hookName {
	case HookName:
		return lsPath(info)
	default:
		return nil, fmt.Errorf("unknown hook name %s", hookName)
	}
}

func lsPath(info component.HookBuildData) (*component.Hook, error) {
	hook := component.Hook{
		RequestVars:      nil,
		RequestResources: nil,
		Resource:         []byte{},
	}

	task := v1alpha1.PipelineTask{
		Name: HookName,
		TaskSpec: &v1alpha1.TaskSpec{
			TaskSpec: v1beta1.TaskSpec{
				Params: []v1beta1.ParamSpec{{Name: "Path"}},
				Steps: []v1beta1.Step{
					{
						Name:   "print-path",
						Image:  "bash:4.4",
						Script: "ls $(params.Path)",
					},
				},
			},
		},
	}

	if path, ok := info.UserVars[varNamePrintPath]; ok {
		task.Params = []v1beta1.Param{
			{
				Name:  "Path",
				Value: *v1beta1.NewArrayOrString(path),
			},
		}
	}

	if imageName, ok := info.UserVars[varNameImageName]; ok {
		task.TaskSpec.TaskSpec.Steps[0].Image = imageName
	}

	taskStr, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}
	appLogger.Debug("task info", "task", taskStr)
	hook.Resource = taskStr
	return &hook, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			string(shared.PluginTypeGRPC): &shared.HookFactoryPlugin{Impl: &report{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     appLogger,
	})
}
