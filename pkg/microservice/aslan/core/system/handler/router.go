/*
Copyright 2021 The KodeRover Authors.

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
	proxy := router.Group("proxy")
	{
		proxy.GET("/config", GetProxyConfig)
	}

	// ---------------------------------------------------------------------------------------
	// 安装脚本管理接口
	// ---------------------------------------------------------------------------------------
	install := router.Group("install")
	{
		install.POST("", CreateInstall)
		install.PUT("", UpdateInstall)
		install.GET("/:name/:version", GetInstall)
		install.GET("", ListInstalls)
		install.PUT("/delete", DeleteInstall)
	}

	// ---------------------------------------------------------------------------------------
	// 代理管理接口
	// ---------------------------------------------------------------------------------------
	proxyManage := router.Group("proxyManage")
	{
		proxyManage.GET("", ListProxies)
		proxyManage.GET("/:id", GetProxy)
		proxyManage.POST("", CreateProxy)
		proxyManage.PUT("/:id", UpdateProxy)
		proxyManage.DELETE("/:id", DeleteProxy)

		proxyManage.POST("/connectionTest", TestConnection)
	}

	registry := router.Group("registry")
	{
		registry.GET("", ListRegistries)
		// 获取默认的镜像仓库配置，用于kodespace CLI调用
		registry.GET("/namespaces/default", GetDefaultRegistryNamespace)
		registry.GET("/namespaces/specific/:id", GetRegistryNamespace)
		registry.GET("/namespaces", ListRegistryNamespaces)
		registry.POST("/namespaces", CreateRegistryNamespace)
		registry.PUT("/namespaces/:id", UpdateRegistryNamespace)

		registry.DELETE("/namespaces/:id", DeleteRegistryNamespace)
		registry.GET("/release/repos", ListAllRepos)
		registry.POST("/images", ListImages)
		registry.GET("/images/repos/:name", ListRepoImages)
	}

	s3storage := router.Group("s3storage")
	{
		s3storage.GET("", ListS3Storage)
		s3storage.POST("", CreateS3Storage)
		s3storage.GET("/:id", GetS3Storage)
		s3storage.PUT("/:id", UpdateS3Storage)
		s3storage.DELETE("/:id", DeleteS3Storage)
		s3storage.POST("/:id/releases/search", ListTars)
	}

	//系统清理缓存
	cleanCache := router.Group("cleanCache")
	{
		cleanCache.POST("/oneClick", CleanImageCache)
		cleanCache.GET("/state", CleanCacheState)
		cleanCache.POST("/cron", SetCron)
	}

	// ---------------------------------------------------------------------------------------
	// Github管理接口
	// ---------------------------------------------------------------------------------------
	github := router.Group("githubApp")
	{
		github.GET("", GetGithubApp)
		github.POST("", CreateGithubApp)
		github.DELETE("/:id", DeleteGithubApp)
	}

	// ---------------------------------------------------------------------------------------
	// jenkins集成接口以及jobs和buildWithParameters接口
	// ---------------------------------------------------------------------------------------
	jenkins := router.Group("jenkins")
	{
		jenkins.GET("/exist", CheckJenkinsIntegration)
		jenkins.POST("/integration", CreateJenkinsIntegration)
		jenkins.GET("/integration", ListJenkinsIntegration)
		jenkins.PUT("/integration/:id", UpdateJenkinsIntegration)
		jenkins.DELETE("/integration/:id", DeleteJenkinsIntegration)
		jenkins.POST("/user/connection", TestJenkinsConnection)
		jenkins.GET("/jobNames/:id", ListJobNames)
		jenkins.GET("/buildArgs/:id/:jobName", ListJobBuildArgs)
	}

	//系统配额
	capacity := router.Group("capacity")
	{
		capacity.POST("", UpdateStrategy)
		capacity.GET("/target/:target", GetStrategy)
		capacity.POST("/gc", GarbageCollection)
		// 清理已被删除的工作流的所有缓存，暂时用于手动调用
		capacity.POST("/clean", CleanCache)
	}

	// workflow concurrency settings
	concurrency := router.Group("concurrency")
	{
		concurrency.GET("/workflow", GetWorkflowConcurrency)
		concurrency.POST("/workflow", UpdateWorkflowConcurrency)
	}

	// default login default login home page settings
	login := router.Group("login")
	{
		login.GET("/default", GetDefaultLogin)
		login.POST("/default", UpdateDefaultLogin)
	}

	// ---------------------------------------------------------------------------------------
	// 自定义镜像管理接口
	// ---------------------------------------------------------------------------------------
	basicImages := router.Group("basicImages")
	{
		basicImages.GET("", ListBasicImages)
		basicImages.GET("/:id", GetBasicImage)
		basicImages.POST("", CreateBasicImage)
		basicImages.PUT("/:id", UpdateBasicImage)
		basicImages.DELETE("/:id", DeleteBasicImage)
	}

	// ---------------------------------------------------------------------------------------
	// helm chart 集成
	// ---------------------------------------------------------------------------------------
	integration := router.Group("helm")
	{
		integration.GET("", ListHelmRepos)
		integration.GET("/public", ListHelmReposPublic)
		integration.POST("", CreateHelmRepo)
		integration.PUT("/:id", UpdateHelmRepo)
		integration.DELETE("/:id", DeleteHelmRepo)
		integration.GET("/:name/index", ListCharts)
	}

	// ---------------------------------------------------------------------------------------
	// ssh私钥管理接口
	// ---------------------------------------------------------------------------------------
	privateKey := router.Group("privateKey")
	{
		privateKey.GET("", ListPrivateKeys)
		privateKey.GET("/internal", ListPrivateKeysInternal)
		privateKey.GET("/:id", GetPrivateKey)
		privateKey.GET("/labels", ListLabels)
		privateKey.POST("", CreatePrivateKey)
		privateKey.POST("/batch", BatchCreatePrivateKey)
		privateKey.PUT("/:id", UpdatePrivateKey)
		privateKey.DELETE("/:id", DeletePrivateKey)
	}

	rsaKey := router.Group("rsaKey")
	{
		rsaKey.GET("publicKey", GetRSAPublicKey)
		rsaKey.GET("decryptedText", GetTextFromEncryptedKey)
	}

	notification := router.Group("notification")
	{
		notification.GET("", PullNotify)
		notification.PUT("/read", ReadNotify)
		notification.POST("/delete", DeleteNotifies)
		notification.POST("/subscribe", UpsertSubscription)
		notification.PUT("/subscribe/:type", UpdateSubscribe)
		notification.DELETE("/unsubscribe/notifytype/:type", Unsubscribe)
		notification.GET("/subscribe", ListSubscriptions)
	}

	announcement := router.Group("announcement")
	{
		announcement.POST("", CreateAnnouncement)
		announcement.PUT("/update", UpdateAnnouncement)
		announcement.GET("/all", PullAllAnnouncement)
		announcement.GET("", PullNotifyAnnouncement)
		announcement.DELETE("/:id", DeleteAnnouncement)
	}

	operation := router.Group("operation")
	{
		operation.GET("", GetOperationLogs)
		operation.POST("", AddSystemOperationLog)
		operation.PUT("/:id", UpdateOperationLog)
	}

	// ---------------------------------------------------------------------------------------
	// system external link
	// ---------------------------------------------------------------------------------------
	externalLink := router.Group("externalLink")
	{
		externalLink.GET("", ListExternalLinks)
		externalLink.POST("", CreateExternalLink)
		externalLink.PUT("/:id", UpdateExternalLink)
		externalLink.DELETE("/:id", DeleteExternalLink)
	}

	// ---------------------------------------------------------------------------------------
	// system custom theme
	// ---------------------------------------------------------------------------------------
	theme := router.Group("theme")
	{
		theme.GET("", GetThemeInfos)
		theme.PUT("", UpdateThemeInfo)
	}

	// ---------------------------------------------------------------------------------------
	// external system API
	// ---------------------------------------------------------------------------------------
	externalSystem := router.Group("external")
	{
		externalSystem.POST("", CreateExternalSystem)
		externalSystem.GET("", ListExternalSystem)
		externalSystem.GET("/:id", GetExternalSystemDetail)
		externalSystem.PUT("/:id", UpdateExternalSystem)
		externalSystem.DELETE("/:id", DeleteExternalSystem)
	}

	// ---------------------------------------------------------------------------------------
	// sonar integration API
	// ---------------------------------------------------------------------------------------
	sonar := router.Group("sonar")
	{
		sonar.POST("/integration", CreateSonarIntegration)
		sonar.PUT("/integration/:id", UpdateSonarIntegration)
		sonar.GET("/integration", ListSonarIntegration)
		sonar.GET("/integration/:id", GetSonarIntegration)
		sonar.DELETE("/integration/:id", DeleteSonarIntegration)
		sonar.POST("/validate", ValidateSonarInformation)
	}

	// ---------------------------------------------------------------------------------------
	// configuration management integration API
	// ---------------------------------------------------------------------------------------
	configuration := router.Group("configuration")
	{
		configuration.GET("", ListConfigurationManagement)
		configuration.POST("", CreateConfigurationManagement)
		configuration.PUT("/:id", UpdateConfigurationManagement)
		configuration.GET("/:id", GetConfigurationManagement)
		configuration.DELETE("/:id", DeleteConfigurationManagement)
		configuration.POST("/validate", ValidateConfigurationManagement)
		configuration.GET("/apollo/:id/app", ListApolloApps)
		configuration.GET("/apollo/:id/:app_id/env", ListApolloEnvAndClusters)
		configuration.GET("/apollo/:id/:app_id/:env/:cluster/namespace", ListApolloNamespaces)
	}

	imapp := router.Group("im_app")
	{
		imapp.GET("", ListIMApp)
		imapp.POST("", CreateIMApp)
		imapp.PUT("/:id", UpdateIMApp)
		imapp.DELETE("/:id", DeleteIMApp)
		imapp.POST("/validate", ValidateIMApp)
	}

	lark := router.Group("lark")
	{
		lark.GET("/:id/department/:department_id", GetLarkDepartment)
		lark.GET("/:id/user", GetLarkUserID)
		lark.POST("/:id/webhook", LarkEventHandler)
	}

	dingtalk := router.Group("dingtalk")
	{
		dingtalk.GET("/:id/department/:department_id", GetDingTalkDepartment)
		dingtalk.GET("/:id/user", GetDingTalkUserID)
		dingtalk.POST("/:ak/webhook", DingTalkEventHandler)
	}

	pm := router.Group("project_management")
	{
		pm.GET("", ListProjectManagement)
		pm.POST("", CreateProjectManagement)
		pm.POST("/validate", Validate)
		pm.PATCH("/:id", UpdateProjectManagement)
		pm.DELETE("/:id", DeleteProjectManagement)
		pm.GET("/jira/project", ListJiraProjects)
		pm.GET("/jira/issue", SearchJiraIssues)
		pm.GET("/jira/issue/jql", SearchJiraProjectIssuesWithJQL)
		pm.GET("/jira/type", GetJiraTypes)
		pm.GET("/jira/status", GetJiraAllStatus)
		pm.POST("/jira/webhook/:workflowName/:hookName", HandleJiraEvent)
		pm.POST("/meego/webhook/:workflowName/:hookName", HandleMeegoEvent)
	}
	// personal dashboard configuration
	dashboard := router.Group("dashboard")
	{
		// dashboard configuration
		dashboard.GET("/settings", GetDashboardConfiguration)
		dashboard.PUT("/settings", CreateOrUpdateDashboardConfiguration)

		// dashboard card API
		dashboard.GET("/workflow/running", GetRunningWorkflow)
		dashboard.GET("/workflow/mine", GetMyWorkflow)
		dashboard.GET("/environment/:name", GetMyEnvironment)
	}

	// get nacos info
	nacos := router.Group("nacos")
	{
		nacos.GET("/:nacosID", ListNacosNamespace)
		nacos.GET("/:nacosID/namespace/:nacosNamespaceID", ListNacosConfig)
	}

	// feishu project management module
	meego := router.Group("meego")
	{
		meego.GET("/projects", GetMeegoProjects)
		meego.GET("/projects/:projectID/work_item/types", GetWorkItemTypeList)
		meego.GET("/projects/:projectID/work_item", ListMeegoWorkItems)
		meego.GET("/projects/:projectID/work_item/:workItemID/transitions", ListAvailableWorkItemTransitions)
	}

	// personal favorite API
	favorite := router.Group("favorite")
	{
		favorite.POST("/:type/:name", CreateFavorite)
		favorite.DELETE("/:type/:name", DeleteFavorite)
	}

	// ---------------------------------------------------------------------------------------
	// webhook config
	// ---------------------------------------------------------------------------------------
	webhook := router.Group("webhook")
	{
		webhook.GET("/config", GetWebhookConfig)
	}
}

type OpenAPIRouter struct{}

func (*OpenAPIRouter) Inject(router *gin.RouterGroup) {
	reg := router.Group("registry")
	{
		reg.POST("", OpenAPICreateRegistry)
		reg.GET("", OpenAPIListRegistry)
	}

	cluster := router.Group("cluster")
	{
		cluster.GET("", OpenAPIListCluster)
	}
}
