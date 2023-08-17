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

package service

import (
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	commonmodels "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	commonrepo "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/mongodb"
	jobctl "github.com/koderover/zadig/pkg/microservice/aslan/core/workflow/service/workflow/job"
	"github.com/koderover/zadig/pkg/setting"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/tool/log"
)

func CreateWorkflowTemplate(userName string, template *commonmodels.WorkflowV4Template, logger *zap.SugaredLogger) error {
	if _, err := commonrepo.NewWorkflowV4TemplateColl().Find(&commonrepo.WorkflowTemplateQueryOption{Name: template.TemplateName}); err == nil {
		errMsg := fmt.Sprintf("工作流模板名称: %s 已存在", template.TemplateName)
		logger.Error(errMsg)
		return e.ErrCreateWorkflowTemplate.AddDesc(errMsg)
	}
	if err := lintWorkflowTemplate(template, logger); err != nil {
		return err
	}
	template.CreatedBy = userName
	template.UpdatedBy = userName

	workflow := &commonmodels.WorkflowV4{
		Stages: template.Stages,
	}

	for _, stage := range template.Stages {
		for _, job := range stage.Jobs {
			if err := jobctl.Instantiate(job, workflow); err != nil {
				logger.Errorf("Failed to instantiate workflow v4 template, error: %v", err)
				return e.ErrCreateWorkflowTemplate.AddErr(err)
			}
		}
	}
	if err := commonrepo.NewWorkflowV4TemplateColl().Create(template); err != nil {
		errMsg := fmt.Sprintf("Failed to create workflow template %s, err: %v", template.TemplateName, err)
		logger.Error(errMsg)
		return e.ErrCreateWorkflowTemplate.AddDesc(errMsg)
	}
	return nil
}

func UpdateWorkflowTemplate(userName string, template *commonmodels.WorkflowV4Template, logger *zap.SugaredLogger) error {
	if _, err := commonrepo.NewWorkflowV4TemplateColl().Find(&commonrepo.WorkflowTemplateQueryOption{ID: template.ID.Hex()}); err != nil {
		errMsg := fmt.Sprintf("workflow template %s not found: %v", template.TemplateName, err)
		logger.Error(errMsg)
		return e.ErrUpdateWorkflowTemplate.AddDesc(errMsg)
	}
	if err := lintWorkflowTemplate(template, logger); err != nil {
		return err
	}
	template.UpdatedBy = userName

	workflow := &commonmodels.WorkflowV4{
		Stages: template.Stages,
	}

	for _, stage := range template.Stages {
		for _, job := range stage.Jobs {
			if err := jobctl.Instantiate(job, workflow); err != nil {
				logger.Errorf("Failed to instantiate workflow v4 template, error: %v", err)
				return e.ErrUpdateWorkflowTemplate.AddErr(err)
			}
		}
	}
	if err := commonrepo.NewWorkflowV4TemplateColl().Update(template); err != nil {
		errMsg := fmt.Sprintf("Failed to update workflow template %s, err: %v", template.TemplateName, err)
		logger.Error(errMsg)
		return e.ErrUpdateWorkflowTemplate.AddDesc(errMsg)
	}
	return nil
}

func ListWorkflowTemplate(category string, excludeBuildIn bool, logger *zap.SugaredLogger) ([]*WorkflowtemplatePreView, error) {
	resp := []*WorkflowtemplatePreView{}
	templates, err := commonrepo.NewWorkflowV4TemplateColl().List(&commonrepo.WorkflowTemplateListOption{Category: category, ExcludeBuildIn: excludeBuildIn})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to list workflow template err: %v", err)
		logger.Error(errMsg)
		return resp, e.ErrListWorkflowTemplate.AddDesc(errMsg)
	}
	for _, template := range templates {
		stages := []string{}
		for _, stage := range template.Stages {
			if stage.Approval != nil && stage.Approval.Enabled {
				stages = append(stages, "人工审批")
			}
			stages = append(stages, stage.Name)
		}

		stageDetails := make([]*WorkflowTemplateStage, 0)
		for _, stage := range template.Stages {
			stageDetail := &WorkflowTemplateStage{
				Name: stage.Name,
				Jobs: make([]*WorkflowTemplateJob, 0),
			}
			for _, job := range stage.Jobs {
				stageDetail.Jobs = append(stageDetail.Jobs, &WorkflowTemplateJob{
					Name:    job.Name,
					JobType: string(job.JobType),
				})
			}
			stageDetails = append(stageDetails, stageDetail)
		}

		resp = append(resp, &WorkflowtemplatePreView{
			ID:           template.ID,
			TemplateName: template.TemplateName,
			UpdateTime:   template.UpdateTime,
			CreateTime:   template.CreateTime,
			CreateBy:     template.CreatedBy,
			UpdateBy:     template.UpdatedBy,
			Stages:       stages,
			StageDetails: stageDetails,
			Description:  template.Description,
			Category:     template.Category,
			BuildIn:      template.BuildIn,
		})
	}
	return resp, nil
}

func GetWorkflowTemplateByID(idStr string, logger *zap.SugaredLogger) (*commonmodels.WorkflowV4Template, error) {
	template, err := commonrepo.NewWorkflowV4TemplateColl().Find(&commonrepo.WorkflowTemplateQueryOption{ID: idStr})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get workflow template err: %v", err)
		logger.Error(errMsg)
		return template, e.ErrGetWorkflowTemplate.AddDesc(errMsg)
	}
	return template, nil
}

func DeleteWorkflowTemplateByID(idStr string, logger *zap.SugaredLogger) error {
	if err := commonrepo.NewWorkflowV4TemplateColl().DeleteByID(idStr); err != nil {
		errMsg := fmt.Sprintf("Failed to delete workflow template err: %v", err)
		logger.Error(errMsg)
		return e.ErrDeleteWorkflowTemplate.AddDesc(errMsg)
	}
	return nil
}

func lintWorkflowTemplate(template *commonmodels.WorkflowV4Template, logger *zap.SugaredLogger) error {
	stageNameMap := make(map[string]bool)
	jobNameMap := make(map[string]string)

	reg, err := regexp.Compile(setting.JobNameRegx)
	if err != nil {
		logger.Errorf("reg compile failed: %v", err)
		return e.ErrLintWorkflowTemplate.AddErr(err)
	}
	for _, stage := range template.Stages {
		if _, ok := stageNameMap[stage.Name]; !ok {
			stageNameMap[stage.Name] = true
		} else {
			logger.Errorf("duplicated stage name: %s", stage.Name)
			return e.ErrLintWorkflowTemplate.AddDesc(fmt.Sprintf("duplicated stage name: %s", stage.Name))
		}
		for _, job := range stage.Jobs {
			if match := reg.MatchString(job.Name); !match {
				logger.Errorf("job name [%s] did not match %s", job.Name, setting.JobNameRegx)
				return e.ErrLintWorkflowTemplate.AddDesc(fmt.Sprintf("job name [%s] did not match %s", job.Name, setting.JobNameRegx))
			}
			if _, ok := jobNameMap[job.Name]; !ok {
				jobNameMap[job.Name] = string(job.JobType)
			} else {
				logger.Errorf("duplicated job name: %s", job.Name)
				return e.ErrLintWorkflowTemplate.AddDesc(fmt.Sprintf("duplicated job name: %s", job.Name))
			}
		}
	}
	return nil
}

func InitWorkflowTemplate() {
	logger := log.SugaredLogger()
	for _, template := range InitWorkflowTemplateInfos() {
		template.CreateTime = time.Now().Unix()
		template.UpdateTime = time.Now().Unix()
		template.CreatedBy = "system"
		template.UpdatedBy = "system"
		template.ConcurrencyLimit = 1
		if err := commonrepo.NewWorkflowV4TemplateColl().UpsertByName(template); err != nil {
			logger.Errorf("update build-in workflow template error: %v", err)
		}
	}
}

func InitWorkflowTemplateInfos() []*commonmodels.WorkflowV4Template {
	buildInWorkflowTemplateInfos := []*commonmodels.WorkflowV4Template{
		// deploy service
		{
			TemplateName: "单环境多服务更新",
			BuildIn:      true,
			Description:  "支持多个服务并行构建、部署过程",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "多环境多服务更新",
			BuildIn:      true,
			Description:  "支持一次构建部署多个环境",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署环境 dev",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy-env-dev",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "镜像分发",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "image-distribute",
							JobType: config.JobZadigDistributeImage,
							Spec: commonmodels.ZadigDistributeImageJobSpec{
								Source:  config.SourceFromJob,
								JobName: "build",
							},
						},
					},
				},
				{
					Name:     "部署环境 qa",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy-env-qa",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "image-distribute",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "上线服务",
			BuildIn:      true,
			Description:  "支持通过工作流上线服务",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage, config.DeployVars, config.DeployConfig},
								Source:         config.SourceRuntime,
							},
						},
					},
				},
				{
					Name:     "测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec:    commonmodels.ZadigTestingJobSpec{},
						},
					},
				},
			},
		},
		{
			TemplateName: "下线服务",
			BuildIn:      true,
			Description:  "支持通过工作流下线服务",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "下线服务",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "offline-service",
							JobType: config.JobOfflineService,
							Spec:    commonmodels.JobTaskOfflineServiceSpec{},
						},
					},
				},
			},
		},

		// test and security
		{
			TemplateName: "API 测试",
			BuildIn:      true,
			Description:  "支持自动化执行服务级别 API 测试",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "API 测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec: commonmodels.ZadigTestingJobSpec{
								TestType: config.ServiceTestType,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "E2E 测试",
			BuildIn:      true,
			Description:  "支持自动化执行产品级别 E2E 测试",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "E2E 测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec: commonmodels.ZadigTestingJobSpec{
								TestType: config.ProductTestType,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "静态代码检测",
			BuildIn:      true,
			Description:  "支持自动化执行代码扫描和代码成分分析",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "代码扫描",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "code-scanning",
							JobType: config.JobZadigScanning,
							Spec:    commonmodels.ZadigScanningJobSpec{},
						},
					},
				},
				{
					Name:     "成分分析",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "code-analyse",
							JobType: config.JobFreestyle,
							Spec:    commonmodels.FreestyleJobSpec{},
						},
					},
				},
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "动态安全检测",
			BuildIn:      true,
			Description:  "支持自动化执行服务动态安全检测",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "动态安全检测",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec:    commonmodels.ZadigTestingJobSpec{},
						},
					},
				},
			},
		},

		// release
		{
			TemplateName: "Chart 实例化部署",
			BuildIn:      true,
			Description:  "支持自动化部署 Chart 仓库中已有的 Chart 到环境中（仅限 K8s Helm Chart 项目）",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "Chart 部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigHelmDeploy,
							Spec:    commonmodels.ZadigHelmChartDeployJobSpec{},
						},
					},
				},
				{
					Name:     "测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec: commonmodels.ZadigTestingJobSpec{
								TestType: config.ProductTestType,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "多阶段灰度发布",
			Description:  "支持自动化执行多阶段的灰度发布，结合人工审批，确保灰度过程可控",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "灰度20%",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "gray-20",
							JobType: config.JobK8sGrayRelease,
							Spec: commonmodels.GrayReleaseJobSpec{
								DeployTimeout: 10,
								GrayScale:     20,
							},
						},
					},
				},
				{
					Name:     "灰度50%",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm release to 50%",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "gray-50",
							JobType: config.JobK8sGrayRelease,
							Spec: commonmodels.GrayReleaseJobSpec{
								FromJob:       "gray-20",
								DeployTimeout: 10,
								GrayScale:     50,
							},
						},
					},
				},
				{
					Name:     "灰度100%",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm release to 100%",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "gray-100",
							JobType: config.JobK8sGrayRelease,
							Spec: commonmodels.GrayReleaseJobSpec{
								FromJob:       "gray-20",
								DeployTimeout: 10,
								GrayScale:     100,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "蓝绿发布",
			Description:  "支持自动化执行蓝绿发布，结合人工审批，确保蓝绿过程可控",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "蓝绿部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "blue-green-deploy",
							JobType: config.JobK8sBlueGreenDeploy,
							Spec: commonmodels.BlueGreenDeployV2JobSpec{
								Version:    "v2",
								Production: true,
							},
						},
					},
				},
				{
					Name:     "检查",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "check",
							JobType: config.JobFreestyle,
							Spec:    commonmodels.FreestyleJobSpec{},
						},
					},
				},
				{
					Name:     "蓝绿发布",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm to release",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "blue-green-release",
							JobType: config.JobK8sBlueGreenRelease,
							Spec: commonmodels.BlueGreenReleaseV2JobSpec{
								FromJob: "blue-green-deploy",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "金丝雀发布",
			Description:  "支持自动化执行金丝雀发布，结合人工审批，确保金丝雀发布过程可控",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "金丝雀部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "canary-deploy",
							JobType: config.JobK8sCanaryDeploy,
							Spec:    commonmodels.CanaryDeployJobSpec{},
						},
					},
				},
				{
					Name:     "检查",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "check",
							JobType: config.JobFreestyle,
							Spec:    commonmodels.FreestyleJobSpec{},
						},
					},
				},
				{
					Name:     "金丝雀发布",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm to release",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "canary-release",
							JobType: config.JobK8sCanaryRelease,
							Spec: commonmodels.CanaryReleaseJobSpec{
								FromJob:        "canary-deploy",
								ReleaseTimeout: 10,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "Isito 发布",
			Description:  "支持  istio 灰度发布，结合人工审批，确保发布过程可控",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "istio 流量 20%",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "istio-20",
							JobType: config.JobIstioRelease,
							Spec: commonmodels.IstioJobSpec{
								First:             true,
								ClusterID:         setting.LocalClusterID,
								Timeout:           10,
								ReplicaPercentage: 100,
								Weight:            20,
							},
						},
					},
				},
				{
					Name:     "istio 流量 60%",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm to release 60%",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "istio-60",
							JobType: config.JobIstioRelease,
							Spec: commonmodels.IstioJobSpec{
								First:             false,
								FromJob:           "istio-20",
								Timeout:           10,
								ReplicaPercentage: 100,
								Weight:            60,
							},
						},
					},
				},
				{
					Name:     "istio 流量 100%",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm to release 100%",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "istio-100",
							JobType: config.JobIstioRelease,
							Spec: commonmodels.IstioJobSpec{
								First:             false,
								FromJob:           "istio-20",
								Timeout:           10,
								ReplicaPercentage: 100,
								Weight:            100,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "MSE 发布",
			Description:  "支持 MSE 发布，结合人工审批，确保发布过程可控",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "MSE 发布任务",
					Parallel: true,
					Approval: &commonmodels.Approval{
						Enabled:     true,
						Description: "Confirm to mse release",
						Type:        config.NativeApproval,
						NativeApproval: &commonmodels.NativeApproval{
							Timeout:         60,
							NeededApprovers: 1,
						},
					},
					Jobs: []*commonmodels.Job{
						{
							Name:    "mse-release",
							JobType: config.JobMseGrayRelease,
							Spec:    commonmodels.MseGrayReleaseJobSpec{},
						},
					},
				},
				{
					Name:     "检查",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "check",
							JobType: config.JobFreestyle,
							Spec:    commonmodels.FreestyleJobSpec{},
						},
					},
				},
			},
		},

		// project cooperation
		{
			TemplateName: "JIRA 问题状态及业务变更",
			BuildIn:      true,
			Description:  "支持自动化执行 JIRA 问题状态变更",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec:    commonmodels.ZadigTestingJobSpec{},
						},
					},
				},
				{
					Name:     "JIRA 问题状态变更",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "jira-update",
							JobType: config.JobJira,
							Spec: commonmodels.JiraJobSpec{
								Source: setting.VariableSourceRuntime,
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "飞书工作项状态及业务变更",
			BuildIn:      true,
			Description:  "支持自动化执行飞书工作项状态变更",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
				{
					Name:     "测试",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "test",
							JobType: config.JobZadigTesting,
							Spec:    commonmodels.ZadigTestingJobSpec{},
						},
					},
				},
				{
					Name:     "飞书工作项变更",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "lark-update",
							JobType: config.JobMeegoTransition,
							Spec:    commonmodels.MeegoTransitionJobSpec{},
						},
					},
				},
			},
		},

		// config update
		{
			TemplateName: "Nacos 配置及业务变更",
			BuildIn:      true,
			Description:  "支持自动化执行 nacos 配置变更",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "Nacos 配置变更",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "nacos-update",
							JobType: config.JobNacos,
							Spec:    commonmodels.NacosJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "Apollo 配置及业务变更",
			BuildIn:      true,
			Description:  "支持自动化执行 apollo 配置变更",
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "Apollo 配置变更",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "apollo-update",
							JobType: config.JobApollo,
							Spec:    commonmodels.ApolloJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},

		// data update
		{
			TemplateName: "MySQL 数据库及业务变更",
			Description:  "支持自动化执行 MySQL 数据变更以及多服务并行构建、部署过程",
			BuildIn:      true,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name: "MySQL 变更",
					Jobs: []*commonmodels.Job{
						{
							Name:    "mysql-update",
							JobType: config.JobPlugin,
							Spec: commonmodels.PluginJobSpec{
								Properties: &commonmodels.JobProperties{
									Timeout:         10,
									ResourceRequest: setting.MinRequest,
								},
								Plugin: &commonmodels.PluginTemplate{
									Name:        "MySQL 数据库变更",
									IsOffical:   true,
									Description: "针对 MySQL 数据库执行 SQL 变量",
									Version:     "v0.0.1",
									Image:       "koderover.tencentcloudcr.com/koderover-public/mysql-runner:v0.0.1",
									Envs: []*commonmodels.Env{
										{
											Name:  "MYSQL_HOST",
											Value: "$(inputs.mysql_host)",
										},
										{
											Name:  "MYSQL_PORT",
											Value: "$(inputs.mysql_port)",
										},
										{
											Name:  "USERNAME",
											Value: "$(inputs.username)",
										},
										{
											Name:  "PASSWORD",
											Value: "$(inputs.password)",
										},
										{
											Name:  "QUERY",
											Value: "$(inputs.query)",
										},
									},
									Inputs: []*commonmodels.Param{
										{
											Name:        "mysql_host",
											Description: "mysql host",
											ParamsType:  "string",
										},
										{
											Name:        "mysql_port",
											Description: "mysql port",
											ParamsType:  "string",
										},
										{
											Name:        "username",
											Description: "mysql username",
											ParamsType:  "string",
										},
										{
											Name:        "password",
											Description: "mysql password",
											ParamsType:  "string",
										},
										{
											Name:        "query",
											Description: "query to be used",
											ParamsType:  "string",
										},
									},
								},
							},
						},
					},
				},
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},
		{
			TemplateName: "DMS 数据变更及服务升级",
			Description:  "支持自动化创建并跟踪 DMS 数据变更工单以及多服务并行构建、部署过程",
			BuildIn:      true,
			Category:     setting.ReleaseWorkflow,
			Stages: []*commonmodels.WorkflowStage{
				{
					Name:     "DMS 数据变更",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "dms-update",
							JobType: config.JobPlugin,
							Spec: commonmodels.PluginJobSpec{
								Properties: &commonmodels.JobProperties{
									Timeout:         60,
									ResourceRequest: "low",
									ResReqSpec:      setting.LowRequestSpec,
								},
								Plugin: &commonmodels.PluginTemplate{
									Name:        "DMS 数据变更工单",
									IsOffical:   true,
									Category:    "",
									Description: "创建并跟踪 DMS 数据变更工单",
									Version:     "v0.0.1",
									Image:       "koderover.tencentcloudcr.com/koderover-public/dms-approval:v0.0.1",
									Envs: []*commonmodels.Env{
										{
											Name:  "AK",
											Value: "$(inputs.AK)",
										},
										{
											Name:  "SK",
											Value: "$(inputs.SK)",
										},
										{
											Name:  "DBS",
											Value: "$(inputs.DBS)",
										},
										{
											Name:  "AFFECT_ROWS",
											Value: "$(inputs.AFFECT_ROWS)",
										},
										{
											Name:  "EXEC_SQL",
											Value: "$(inputs.EXEC_SQL)",
										},
										{
											Name:  "COMMENT",
											Value: "$(inputs.COMMENT)",
										},
									},
									Inputs: []*commonmodels.Param{
										{
											Name:        "AK",
											Description: "aliyun account AK",
											ParamsType:  "string",
										},
										{
											Name:        "SK",
											Description: "aliyun account AK",
											ParamsType:  "string",
										},
										{
											Name:        "DBS",
											Description: "databases to operate, separated by commas, eg: test@127.0.0.1:3306,test@127.0.0.1:3306",
											ParamsType:  "string",
										},
										{
											Name:        "AFFECT_ROWS",
											Description: "affect rows",
											ParamsType:  "string",
										},
										{
											Name:        "EXEC_SQL",
											Description: "sql to exexute",
											ParamsType:  "text",
										},
										{
											Name:        "COMMENT",
											Description: "comment message",
											ParamsType:  "string",
										},
									},
								},
							},
						},
					},
				},
				{
					Name:     "构建",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "build",
							JobType: config.JobZadigBuild,
							Spec:    commonmodels.ZadigBuildJobSpec{},
						},
					},
				},
				{
					Name:     "部署",
					Parallel: true,
					Jobs: []*commonmodels.Job{
						{
							Name:    "deploy",
							JobType: config.JobZadigDeploy,
							Spec: commonmodels.ZadigDeployJobSpec{
								DeployContents: []config.DeployContent{config.DeployImage},
								Source:         config.SourceFromJob,
								JobName:        "build",
							},
						},
					},
				},
			},
		},
	}
	return buildInWorkflowTemplateInfos
}
