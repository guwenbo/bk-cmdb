/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */
package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/emicklei/go-restful"

	"configcenter/src/common"
	"configcenter/src/common/auditoplog"
	"configcenter/src/common/blog"
	meta "configcenter/src/common/metadata"
	"configcenter/src/common/util"
)

func (ps *ProcServer) BindModuleProcess(req *restful.Request, resp *restful.Response) {
	srvData := ps.newSrvComm(req.Request.Header)

	ownerID := srvData.ownerID
	defErr := srvData.ccErr

	pathParams := req.PathParameters()
	appIDStr := pathParams[common.BKAppIDField]
	appID, _ := strconv.Atoi(appIDStr)
	procIDStr := pathParams[common.BKProcessIDField]
	procID, _ := strconv.Atoi(procIDStr)
	moduleName := pathParams[common.BKModuleNameField]
	params := make([]interface{}, 0)
	cell := make(map[string]interface{})
	cell[common.BKAppIDField] = appID
	cell[common.BKProcessIDField] = procID
	cell[common.BKModuleNameField] = moduleName
	cell[common.BKOwnerIDField] = ownerID
	params = append(params, cell)

	// TODO use change use chan, process model trigger point
	// if err := ps.createProcInstanceModel(appIDStr, procIDStr, moduleName, ownerID, &sourceAPI.ForwardParam{Header:req.Request.Header}); err != nil {
	//     blog.Errorf("fail to create process instance model. err: %v", err)
	//     resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg:defErr.Error(common.CCErrProcBindToMoudleFaile)})
	//     return
	// }

	ret, err := ps.CoreAPI.ProcController().CreateProc2Module(srvData.ctx, srvData.header, params)
	if nil != err {
		blog.Errorf("BindModuleProcess CreateProc2Module http do  error.  err:%s, input:%+v,rid:%s", err.Error(), params, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.Error(common.CCErrCommHTTPDoRequestFailed)})
		return
	}
	if !ret.Result {
		blog.Errorf("BindModuleProcess CreateProc2Module http reply  error. err code:%d err msg:%s, input:%+v,rid:%s", ret.Code, ret.Result, params, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.New(ret.Code, ret.ErrMsg)})
		return
	}

	// save operation log
	log := common.KvMap{common.BKOpDescField: fmt.Sprintf("bind module [%s]", moduleName), common.BKOpTypeField: auditoplog.AuditOpTypeAdd, "inst_id": procID, common.BKContentField: meta.Content{}}
	ps.CoreAPI.AuditController().AddProcLog(srvData.ctx, ownerID, appIDStr, srvData.user, srvData.header, log)

	resp.WriteEntity(meta.NewSuccessResp(nil))
}

func (ps *ProcServer) DeleteModuleProcessBind(req *restful.Request, resp *restful.Response) {
	srvData := ps.newSrvComm(req.Request.Header)
	defErr := srvData.ccErr

	pathParams := req.PathParameters()
	appIDStr := pathParams[common.BKAppIDField]
	appID, _ := strconv.Atoi(appIDStr)
	procIDStr := pathParams[common.BKProcessIDField]
	procID, _ := strconv.Atoi(procIDStr)
	moduleName := pathParams[common.BKModuleNameField]
	cell := make(map[string]interface{})
	cell[common.BKAppIDField] = appID
	cell[common.BKProcessIDField] = procID
	cell[common.BKModuleNameField] = moduleName

	if err := srvData.lgc.DeleteProcInstanceModel(srvData.ctx, appIDStr, procIDStr, moduleName); err != nil {
		blog.Errorf("DeleteModuleProcessBind DeleteProcInstanceModel %v,input:%+v,rid:%s", err, cell, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: err})
		return
	}

	ret, err := ps.CoreAPI.ProcController().DeleteProc2Module(srvData.ctx, srvData.header, cell)
	if nil != err {
		blog.Errorf("DeleteModuleProcessBind DeleteProc2Module http do error.  err:%s, input:%+v,rid:%s", err.Error(), cell, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.Error(common.CCErrCommHTTPDoRequestFailed)})
		return
	}
	if !ret.Result {
		blog.Errorf("DeleteModuleProcessBind DeleteProc2Module http reply  error. err code:%d err msg:%s, input:%+v,rid:%s", ret.Code, ret.Result, cell, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.New(ret.Code, ret.ErrMsg)})
		return
	}

	// save operation log
	log := common.KvMap{common.BKOpDescField: fmt.Sprintf("unbind module [%s]", moduleName), common.BKOpTypeField: auditoplog.AuditOpTypeAdd, "inst_id": procID, common.BKContentField: meta.Content{}}
	ps.CoreAPI.AuditController().AddProcLog(srvData.ctx, srvData.ownerID, appIDStr, srvData.user, srvData.header, log)

	resp.WriteEntity(meta.NewSuccessResp(nil))
}

func (ps *ProcServer) GetProcessBindModule(req *restful.Request, resp *restful.Response) {
	srvData := ps.newSrvComm(req.Request.Header)
	defErr := srvData.ccErr

	pathParams := req.PathParameters()
	appIDStr := pathParams[common.BKAppIDField]
	appID, errAppID := strconv.Atoi(appIDStr)
	procIDStr := pathParams[common.BKProcessIDField]
	procID, errProcID := strconv.Atoi(procIDStr)

	if nil != errAppID {
		blog.Errorf("GetProcessBindModule application id %s not integer,rid:%s", appIDStr, srvData.rid)
		resp.WriteError(http.StatusBadRequest, &meta.RespError{Msg: defErr.Errorf(common.CCErrCommParamsNeedInt, common.BKAppIDField)})
		return
	}
	if nil != errProcID {
		blog.Errorf("GetProcessBindModule process id %s not integer,rid:%s", procIDStr, srvData.rid)
		resp.WriteError(http.StatusBadRequest, &meta.RespError{Msg: defErr.Errorf(common.CCErrCommParamsNeedInt, common.BKProcessIDField)})
		return
	}
	// search object instance
	condition := make(map[string]interface{})
	condition[common.BKAppIDField] = appID
	input := new(meta.QueryInput)
	input.Condition = condition

	objRet, err := ps.CoreAPI.ObjectController().Instance().SearchObjects(srvData.ctx, common.BKInnerObjIDModule, srvData.header, input)
	if nil != err {
		blog.Errorf("GetProcessBindModule SearchObjects http do error.  err:%s, input:%+v,rid:%s", err.Error(), input, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.Error(common.CCErrCommHTTPDoRequestFailed)})
		return
	}
	if !objRet.Result {
		blog.Errorf("GetProcessBindModule SearchObjects http reply  error. err code:%d err msg:%s, input:%+v,rid:%s", objRet.Code, objRet.Result, input, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.New(objRet.Code, objRet.ErrMsg)})
		return
	}

	condition[common.BKProcessIDField] = procID
	// get process by module
	p2mRet, err := ps.CoreAPI.ProcController().GetProc2Module(context.Background(), req.Request.Header, condition)
	if nil != err {
		blog.Errorf("GetProcessBindModule GetProc2Module http do error.  err:%s, input:%+v,rid:%s", err.Error(), condition, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.Error(common.CCErrCommHTTPDoRequestFailed)})
		return
	}
	if !objRet.Result {
		blog.Errorf("GetProcessBindModule GetProc2Module http reply  error. err code:%d err msg:%s, input:%+v,rid:%s", p2mRet.Code, objRet.Result, condition, srvData.rid)
		resp.WriteError(http.StatusInternalServerError, &meta.RespError{Msg: defErr.New(p2mRet.Code, p2mRet.ErrMsg)})
		return
	}

	disModuleNameArr := make([]string, 0)
	for _, moduleInfo := range objRet.Data.Info {
		if !util.InArray(moduleInfo[common.BKModuleNameField], disModuleNameArr) {
			moduleName, ok := moduleInfo[common.BKModuleNameField].(string)
			if !ok {
				continue
			}
			if moduleInfo.Exists(common.BKDefaultField) {
				isDefault64, err := moduleInfo.Int64(common.BKDefaultField)
				if nil != err {
					blog.Warnf("get module default error:%s,input:%+v,rid:%s", err.Error(), input, srvData.rid)
				} else {
					if 0 != isDefault64 {
						continue
					}
				}
				disModuleNameArr = append(disModuleNameArr, moduleName)
			} else {
				blog.Errorf("ApplicationID %d  module name %s not found default field,rid:%s", appID, moduleName, srvData.rid)
			}

		}
	}

	result := make([]interface{}, 0)
	for _, disModuleName := range disModuleNameArr {
		num := 0
		isBind := 0
		for _, moduleInfo := range objRet.Data.Info {
			moduleName, ok := moduleInfo[common.BKModuleNameField].(string)
			if !ok {
				continue
			}

			if disModuleName == moduleName {
				num++
			}
		}

		for _, procModule := range p2mRet.Data {
			if disModuleName == procModule.ModuleName {
				isBind = 1
				break
			}
		}

		data := make(map[string]interface{})
		data[common.BKModuleNameField] = disModuleName
		data["set_num"] = num
		data["is_bind"] = isBind
		result = append(result, data)
	}

	resp.WriteEntity(meta.NewSuccessResp(result))
}
