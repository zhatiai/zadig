package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	svcservice "github.com/koderover/zadig/pkg/microservice/aslan/core/service/service"
	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
	e "github.com/koderover/zadig/pkg/tool/errors"
)

func LoadServiceFromYamlTemplateOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	req := new(svcservice.OpenAPILoadServiceFromYamlTemplateReq)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}
	req.Production = false

	if err := req.Validate(); err != nil {
		ctx.Err = err
		return
	}

	bs, _ := json.Marshal(req)
	internalhandler.InsertOperationLog(c, ctx.UserName+"(OpenAPI)", req.ProjectKey, "新增", "项目管理-测试服务", fmt.Sprintf("服务名称:%s", req.ServiceName), string(bs), ctx.Logger)

	ctx.Err = svcservice.OpenAPILoadServiceFromYamlTemplate(ctx.UserName, req, false, ctx.Logger)
}

func LoadProductionServiceFromYamlTemplateOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	req := new(svcservice.OpenAPILoadServiceFromYamlTemplateReq)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}
	req.Production = true

	if err := req.Validate(); err != nil {
		ctx.Err = err
		return
	}

	bs, _ := json.Marshal(req)
	internalhandler.InsertOperationLog(c, ctx.UserName+"(OpenAPI)", req.ProjectKey, "新增", "项目管理-生产服务", fmt.Sprintf("服务名称:%s", req.ServiceName), string(bs), ctx.Logger)

	ctx.Err = svcservice.OpenAPILoadServiceFromYamlTemplate(ctx.UserName, req, false, ctx.Logger)
}

func CreateRawYamlServicesOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectKey := c.Query("projectName")
	if projectKey == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid project name")
		return
	}

	req := new(svcservice.OpenAPICreateYamlServiceReq)
	data, err := c.GetRawData()
	if err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid create raw yaml service json args")
		return
	}
	if err = json.Unmarshal(data, req); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid create raw yaml service json args")
		return
	}

	internalhandler.InsertOperationLog(c, ctx.UserName+"(openapi)", projectKey, "新增", "项目管理-测试服务", fmt.Sprintf("服务名称:%s", req.ServiceName), string(data), ctx.Logger)
	ctx.Err = svcservice.CreateRawYamlServicesOpenAPI(ctx.UserName, projectKey, req, ctx.Logger)
}

func CreateRawProductionYamlServicesOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectKey := c.Query("projectName")
	if projectKey == "" {
		ctx.Err = fmt.Errorf("projectName cannot be empty")
		return
	}

	req := new(svcservice.OpenAPICreateYamlServiceReq)
	data, err := c.GetRawData()
	if err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid create raw yaml service json args")
		return
	}
	if err = json.Unmarshal(data, req); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid create raw yaml service json args")
		return
	}
	req.Production = true

	internalhandler.InsertOperationLog(c, ctx.UserName+"(openapi)", projectKey, "新增", "项目管理-生产服务", fmt.Sprintf("服务名称:%s", req.ServiceName), string(data), ctx.Logger)
	ctx.Err = svcservice.CreateRawYamlServicesOpenAPI(ctx.UserName, projectKey, req, ctx.Logger)
}

func UpdateServiceConfigOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(svcservice.OpenAPIUpdateServiceConfigArgs)
	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid update service config json args")
		return
	}
	args.ProjectName = c.Query("projectName")
	args.ServiceName = c.Param("name")
	if err := args.Validate(); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	bs, _ := json.Marshal(args)
	internalhandler.InsertOperationLog(c, ctx.UserName, args.ProjectName, "(OpenAPI)"+"更新测试服务配置", "项目管理-服务", fmt.Sprintf("服务名称:%s", args.ServiceName), string(bs), ctx.Logger)

	ctx.Err = svcservice.OpenAPIUpdateServiceConfig(ctx.UserName, args, ctx.Logger)
}

func UpdateProductionServiceConfigOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(svcservice.OpenAPIUpdateServiceConfigArgs)
	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc("invalid update service config json args")
		return
	}
	args.ProjectName = c.Query("projectName")
	args.ServiceName = c.Param("name")
	if err := args.Validate(); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	bs, _ := json.Marshal(args)
	internalhandler.InsertOperationLog(c, ctx.UserName, args.ProjectName, "(OpenAPI)"+"更新生产服务配置", "项目管理-服务", fmt.Sprintf("服务名称:%s", args.ServiceName), string(bs), ctx.Logger)

	ctx.Err = svcservice.OpenAPIProductionUpdateServiceConfig(ctx.UserName, args, ctx.Logger)
}

func UpdateServiceVariableOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	req := new(svcservice.OpenAPIUpdateServiceVariableRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}

	projectName := c.Query("projectName")
	if projectName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}
	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}

	internalhandler.InsertOperationLog(c, ctx.UserName, projectName, "(OpenAPI)"+"更新测试服务变量", "项目管理-服务", fmt.Sprintf("服务名称:%s", serviceName), "", ctx.Logger)

	ctx.Err = svcservice.OpenAPIUpdateServiceVariable(ctx.UserName, projectName, serviceName, req, ctx.Logger)
}

func UpdateProductionServiceVariableOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	req := new(svcservice.OpenAPIUpdateServiceVariableRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		ctx.Err = err
		return
	}

	projectName := c.Query("projectName")
	if projectName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}
	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}

	internalhandler.InsertOperationLog(c, ctx.UserName, projectName, "(OpenAPI)"+"更新生产服务变量", "项目管理-服务", fmt.Sprintf("服务名称:%s", serviceName), "", ctx.Logger)

	ctx.Err = svcservice.OpenAPIUpdateProductionServiceVariable(ctx.UserName, projectName, serviceName, req, ctx.Logger)
}

func DeleteYamlServicesOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectName := c.Query("projectName")
	if projectName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}
	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}

	internalhandler.InsertOperationLog(c, ctx.UserName, projectName, "OpenAPI"+"删除", "项目管理-测试服务", serviceName, "", ctx.Logger)
	ctx.Err = svcservice.DeleteServiceTemplate(serviceName, "k8s", projectName, c.DefaultQuery("isEnvTemplate", "true"), "private", ctx.Logger)
}

func DeleteProductionServicesOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectName := c.Query("projectName")
	if projectName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}
	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}
	internalhandler.InsertOperationLog(c, ctx.UserName, projectName, "OpenAPI"+"删除", "项目管理-生产服务", serviceName, "", ctx.Logger)

	ctx.Err = svcservice.DeleteProductionServiceTemplate(serviceName, projectName, ctx.Logger)
}

func GetYamlServiceOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectKey := c.Query("projectName")
	if projectKey == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}

	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}

	ctx.Resp, ctx.Err = svcservice.OpenAPIGetYamlService(projectKey, serviceName, ctx.Logger)
}

func GetProductionYamlServiceOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectKey := c.Query("projectName")
	if projectKey == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}

	serviceName := c.Param("name")
	if serviceName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("serviceName cannot be empty")
		return
	}

	ctx.Resp, ctx.Err = svcservice.GetProductionYamlServiceOpenAPI(projectKey, serviceName, ctx.Logger)
}

func ListYamlServicesOpenAPI(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	projectName := c.Query("projectName")
	if projectName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("projectName cannot be empty")
		return
	}

	resp, err := svcservice.ListServiceTemplateOpenAPI(projectName, ctx.Logger)
	if err != nil {
		ctx.Err = err
		return
	}
	ctx.Resp = resp
}
