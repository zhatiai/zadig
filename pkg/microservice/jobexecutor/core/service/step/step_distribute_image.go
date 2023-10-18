/*
Copyright 2022 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package step

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/koderover/zadig/pkg/microservice/jobexecutor/core/service/meta"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/types/ref"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/koderover/zadig/pkg/tool/log"
	"github.com/koderover/zadig/pkg/types/step"
)

type DistributeImageStep struct {
	spec           *step.StepImageDistributeSpec
	envs           []string
	secretEnvs     []string
	dirs           *meta.ExecutorWorkDirs
	logger         *zap.SugaredLogger
	infrastructure string
}

func NewDistributeImageStep(metaData *meta.JobMetaData, logger *zap.SugaredLogger) (*DistributeImageStep, error) {
	distributeImageStep := &DistributeImageStep{dirs: metaData.Dirs, envs: metaData.Envs, secretEnvs: metaData.SecretEnvs, logger: logger, infrastructure: metaData.Infrastructure}
	spec := metaData.Step.Spec
	yamlBytes, err := yaml.Marshal(spec)
	if err != nil {
		return distributeImageStep, fmt.Errorf("marshal spec %+v failed", spec)
	}
	if err := yaml.Unmarshal(yamlBytes, &distributeImageStep.spec); err != nil {
		return distributeImageStep, fmt.Errorf("unmarshal spec %s to shell spec failed", yamlBytes)
	}
	return distributeImageStep, nil
}

func (s *DistributeImageStep) Run(ctx context.Context) error {
	s.logger.Info("Start distribute images.")
	if s.spec.SourceRegistry == nil || s.spec.TargetRegistry == nil {
		return errors.New("image registry infos are missing")
	}
	hostsOpt := regclient.WithConfigHosts([]config.Host{getDockerHost(s.spec.SourceRegistry), getDockerHost(s.spec.TargetRegistry)})
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	client := regclient.New(hostsOpt, regclient.WithLog(logger))

	errList := new(multierror.Error)
	wg := sync.WaitGroup{}
	for _, target := range s.spec.DistributeTarget {
		wg.Add(1)
		go func(target *step.DistributeTaskTarget) {
			defer wg.Done()
			if err := copyImage(target, client); err != nil {
				errList = multierror.Append(errList, err)
			}
		}(target)
	}
	wg.Wait()
	if err := errList.ErrorOrNil(); err != nil {
		return fmt.Errorf("copy images error: %v", err)
	}
	s.logger.Info("Finish distribute images.")
	return nil
}

func copyImage(target *step.DistributeTaskTarget, client *regclient.RegClient) error {
	sourceRef, err := ref.New(target.SourceImage)
	if err != nil {
		errMsg := fmt.Sprintf("parse source image: %s error: %v", target.SourceImage, err)
		return errors.New(errMsg)
	}
	targetRef, err := ref.New(target.TargetImage)
	if err != nil {
		errMsg := fmt.Sprintf("parse target image: %s error: %v", target.TargetImage, err)
		return errors.New(errMsg)
	}
	if err := client.ImageCopy(context.Background(), sourceRef, targetRef); err != nil {
		errMsg := fmt.Sprintf("copy image failed: %v", err)
		return errors.New(errMsg)
	}
	log.Infof("copy image from [%s] to [%s] succeed", target.SourceImage, target.TargetImage)
	return nil
}

func getDockerHost(reg *step.RegistryNamespace) config.Host {
	host := config.HostNewName(reg.RegAddr)
	host.User = reg.AccessKey
	host.Pass = reg.SecretKey
	host.RegCert = reg.TLSCert
	if !reg.TLSEnabled {
		host.TLS = config.TLSInsecure
	}
	if strings.HasPrefix(reg.RegAddr, "http://") {
		host.TLS = config.TLSDisabled
	}
	return *host
}
