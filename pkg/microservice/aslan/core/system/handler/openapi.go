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
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/system/service"
	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/tool/log"
)

func OpenAPICreateRegistry(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(service.OpenAPICreateRegistryReq)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("OpenAPICreateRegistry c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("CreateRegistryNamespace json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.UserName+"(openAPI)", "", "新增", "系统设置-Registry", fmt.Sprintf("提供商:%s,Namespace:%s", args.Provider, args.Namespace), string(data), ctx.Logger)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	err = args.Validate()
	if err != nil {
		ctx.Err = err
		return
	}

	ctx.Err = service.OpenAPICreateRegistry(ctx.UserName, args, ctx.Logger)
}

func OpenAPIListRegistry(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	registry, err := service.ListRegistries(ctx.Logger)
	if err != nil {
		ctx.Err = err
		return
	}

	resp := make([]*service.OpenAPIRegistry, 0)
	for _, reg := range registry {
		resp = append(resp, &service.OpenAPIRegistry{
			ID:        reg.ID.Hex(),
			Address:   reg.RegAddr,
			Provider:  config.RegistryProvider(reg.RegProvider),
			Region:    reg.Region,
			IsDefault: reg.IsDefault,
			Namespace: reg.Namespace,
		})
	}
	ctx.Resp = resp
}

func OpenAPIListCluster(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	ctx.Resp, ctx.Err = service.OpenAPIListCluster(c.Query("projectName"), ctx.Logger)
}
