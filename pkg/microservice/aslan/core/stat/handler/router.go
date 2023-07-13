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

package handler

import (
	"github.com/gin-gonic/gin"
)

type Router struct{}

func (*Router) Inject(router *gin.RouterGroup) {
	dashboard := router.Group("dashboard")
	{
		dashboard.GET("/overview", GetOverviewStat)
		dashboard.GET("/build", GetBuildStat)
		dashboard.GET("/deploy", GetDeployStat)
		dashboard.GET("/test", GetTestDashboard)
	}

	quality := router.Group("quality")
	{
		//buildStat
		quality.POST("/initBuildStat", GetAllPipelineTask)
		quality.POST("/buildAverageMeasure", GetBuildDailyAverageMeasure)
		quality.POST("/buildDailyMeasure", GetBuildDailyMeasure)
		quality.POST("/buildHealthMeasure", GetBuildHealthMeasure)
		quality.POST("/buildLatestTenMeasure", GetLatestTenBuildMeasure)
		quality.POST("/buildTenDurationMeasure", GetTenDurationMeasure)
		quality.POST("/buildTrend", GetBuildTrendMeasure)
		//testStat
		quality.POST("/initTestStat", InitTestStat)
		quality.POST("/testAverageMeasure", GetTestAverageMeasure)
		quality.POST("/testCaseMeasure", GetTestCaseMeasure)
		quality.POST("/testDeliveryDeploy", GetTestDeliveryDeployMeasure)
		quality.POST("/testHealthMeasure", GetTestHealthMeasure)
		quality.POST("/testTrend", GetTestTrendMeasure)
		//deployStat
		quality.POST("/initDeployStat", InitDeployStat)
		quality.POST("/pipelineHealthMeasure", GetPipelineHealthMeasure)
		quality.POST("/deployHealthMeasure", GetDeployHealthMeasure)
		quality.POST("/deployWeeklyMeasure", GetDeployWeeklyMeasure)
		quality.POST("/deployTopFiveHigherMeasure", GetDeployTopFiveHigherMeasure)
		quality.POST("/deployTopFiveFailureMeasure", GetDeployTopFiveFailureMeasure)
	}

	// v2 api, mainly for enterprise statistics
	v2 := router.Group("v2")
	{
		v2.GET("/config", ListStatDashboardConfigs)
		v2.POST("/config", CreateStatDashboardConfig)
		v2.PUT("/config/:id", UpdateStatDashboardConfig)
		v2.DELETE("/config/:id", DeleteStatDashboardConfig)
		v2.GET("/dashboard", GetStatsDashboard)
		v2.GET("/dashboard/general", GetStatsDashboardGeneralData)
		// ai api TODO: consider api call auth
		v2.POST("/ai/analysis", GetAIStatsAnalysis)
		v2.GET("/ai/analysis/prompt", GetAIStatsAnalysisPrompts)
		v2.GET("/ai/overview", GetProjectsOverview)
		v2.GET("/ai/build/trend", GetCurrently30DayBuildTrend)
		v2.GET("/ai/radar", GetEfficiencyRadar)
		v2.GET("/ai/attention", GetMonthAttention)
		v2.GET("/ai/requirement/period", GetRequirementDevDepPeriod)
	}
}

type OpenAPIRouter struct{}

func (*OpenAPIRouter) Inject(router *gin.RouterGroup) {
	dashboard := router.Group("")
	{
		dashboard.GET("/overview", GetOverviewStat)
		dashboard.GET("/build", GetBuildStatForOpenAPI)
		dashboard.GET("/deploy", GetDeployStatsOpenAPI)
		dashboard.GET("/test", GetTestStatOpenAPI)
	}

	// enterprise statistics OpenAPI
	v2 := router.Group("/v2")
	{
		v2.GET("/release", GetReleaseStatOpenAPI)
	}
}
