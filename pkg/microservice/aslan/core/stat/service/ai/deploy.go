package ai

import (
	"encoding/json"
	"fmt"
	"time"

	service2 "github.com/koderover/zadig/pkg/microservice/aslan/core/stat/service"
	"go.uber.org/zap"
)

type DeployData struct {
	Description string         `json:"description"`
	Details     *DeployDetails `json:"deploy_details"`
}

type DeployDetails struct {
	DeployTotal                     int           `json:"deploy_total"`
	DeploySuccessTotal              int           `json:"deploy_success_total"`
	DeployFailureTotal              int           `json:"deploy_failure_total"`
	DeployTotalDuration             int64         `json:"deploy_total_duration"`
	DeployHealthMeasureData         *DeployDetail `json:"deploy_health_measure_data"`
	DeployWeeklyMeasureData         *DeployDetail `json:"deploy_weekly_measure_data"`
	DeployTopFiveHigherMeasureData  *DeployDetail `json:"deploy_top_five_higher_measure_data"`
	DeployTopFiveFailureMeasureData *DeployDetail `json:"deploy_top_five_failure_measure_data"`
}

type DeployDetail struct {
	Description string `json:"description"`
	Details     string `json:"details"`
}

func getDeployData(project string, startTime, endTime int64, log *zap.SugaredLogger) (*DeployData, error) {
	log.Infof("=====> Start to get deploy data for project %s", project)
	deploy := &DeployData{
		Description: fmt.Sprintf("%s项目在%s到%s期间部署相关数据，包括部署总次数，部署通过次数，部署失败次数, 部署周趋势数据，部署每日数据", project, time.Unix(startTime, 0).Format("2006-01-02"), time.Unix(endTime, 0).Format("2006-01-02")),
		Details:     &DeployDetails{},
	}
	// get deploy data from mongo
	deployJobList, err := service2.GetProjectDeployStat(startTime, endTime, project)
	if err != nil {
		return deploy, err
	}

	deploy.Details.DeployTotal = deployJobList.Total
	deploy.Details.DeploySuccessTotal = deployJobList.Success
	deploy.Details.DeployFailureTotal = deployJobList.Failure
	deploy.Details.DeployTotalDuration = int64(deployJobList.Duration)

	deploy.Details.DeployWeeklyMeasureData = &DeployDetail{}
	getDeployWeeklyMeasure(project, startTime, endTime, deploy.Details.DeployWeeklyMeasureData, log)

	// TODO: DeployTopFiveHigherMeasureData is old method, need to be upgrade
	//deploy.Details.DeployTopFiveHigherMeasureData = &DeployDetail{}
	//getDeployTopFiveHigherMeasure(project, startTime, endTime, deploy.Details.DeployTopFiveHigherMeasureData, log)

	// TODO: DeployTopFiveFailureMeasureData is old method, need to be upgrade
	//deploy.Details.DeployTopFiveFailureMeasureData = &DeployDetail{}
	//getDeployTopFiveFailureMeasure(project, startTime, endTime, deploy.Details.DeployTopFiveFailureMeasureData, log)

	return deploy, nil
}

func getDeployHealthMeasure(project string, startTime, endTime int64, detail *DeployDetail, log *zap.SugaredLogger) {
	// get deploy health measure data
	DeployHealthMeasure, err := service2.GetDeployHealthMeasure(startTime, endTime, []string{project}, log)
	if err != nil {
		log.Errorf("Failed to get deploy health measure data, the error is: %+v", err)
	}

	health, err := json.Marshal(DeployHealthMeasure)
	if err != nil {
		log.Errorf("Failed to marshal deploy health measure data, the error is: %+v", err)
	}
	detail.Description = "服务部署健康度"
	detail.Details = string(health)
}

func getDeployWeeklyMeasure(project string, startTime, endTime int64, detail *DeployDetail, log *zap.SugaredLogger) {
	// get deploy weekly measure data
	DeployWeeklyMeasure, err := service2.GetProjectsWeeklyDeployStat(startTime, endTime, []string{project})
	if err != nil {
		log.Errorf("Failed to get deploy weekly measure data, the error is: %+v", err)
	}

	weekly, err := json.Marshal(DeployWeeklyMeasure)
	if err != nil {
		log.Errorf("Failed to marshal deploy weekly measure data, the error is: %+v", err)
	}
	detail.Description = fmt.Sprintf("%s项目在%s到%s期间的部署趋势数据，统计一周内总的部署次数，部署成功次数，部署失败次数，部署平均耗时等数据", project, time.Unix(startTime, 0).Format("2006-01-02"), time.Unix(endTime, 0).Format("2006-01-02"))
	log.Infof("%s:\n%s", detail.Description, string(weekly))
	detail.Details = string(weekly)
}

func getDeployTopFiveHigherMeasure(project string, startTime, endTime int64, detail *DeployDetail, log *zap.SugaredLogger) {
	// get deploy top five higher measure data
	DeployTopFiveHigherMeasure, err := service2.GetDeployTopFiveHigherMeasure(startTime, endTime, []string{project}, log)
	if err != nil {
		log.Errorf("Failed to get deploy top five higher measure data, the error is: %+v", err)
	}

	higher, err := json.Marshal(DeployTopFiveHigherMeasure)
	if err != nil {
		log.Errorf("Failed to marshal deploy top five higher measure data, the error is: %+v", err)
	}
	detail.Description = fmt.Sprintf("%s项目在%s到%s期间的部署耗时Top5数据", project, time.Unix(startTime, 0).Format("2006-01-02"), time.Unix(endTime, 0).Format("2006-01-02"))
	log.Infof("%s:\n%s", detail.Description, string(higher))
	detail.Details = string(higher)
}

func getDeployTopFiveFailureMeasure(project string, startTime, endTime int64, detail *DeployDetail, log *zap.SugaredLogger) {
	// get deploy top five failure measure data
	DeployTopFiveFailureMeasure, err := service2.GetDeployTopFiveFailureMeasure(startTime, endTime, []string{project}, log)
	if err != nil {
		log.Errorf("Failed to get deploy top five failure measure data, the error is: %+v", err)
	}

	failure, err := json.Marshal(DeployTopFiveFailureMeasure)
	if err != nil {
		log.Errorf("Failed to marshal deploy top five failure measure data, the error is: %+v", err)
	}
	detail.Description = fmt.Sprintf("%s项目在%s-%s期间的部署失败Top5数据", project, time.Unix(startTime, 0).Format("2006-01-02"), time.Unix(endTime, 0).Format("2006-01-02"))
	log.Infof("%s:\n%s", detail.Description, string(failure))
	detail.Details = string(failure)
}
