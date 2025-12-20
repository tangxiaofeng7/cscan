package logic

import (
	"context"
	"fmt"
	"strings"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

// ==================== 标签映射 ====================

type TagMappingListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTagMappingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagMappingListLogic {
	return &TagMappingListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagMappingListLogic) TagMappingList() (resp *types.TagMappingListResp, err error) {
	docs, err := l.svcCtx.TagMappingModel.FindAll(l.ctx)
	if err != nil {
		return &types.TagMappingListResp{Code: 500, Msg: "查询失败"}, nil
	}

	list := make([]types.TagMapping, 0, len(docs))
	for _, doc := range docs {
		list = append(list, types.TagMapping{
			Id:          doc.Id.Hex(),
			AppName:     doc.AppName,
			NucleiTags:  doc.NucleiTags,
			Description: doc.Description,
			Enabled:     doc.Enabled,
			CreateTime:  doc.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.TagMappingListResp{
		Code: 0,
		Msg:  "success",
		List: list,
	}, nil
}

type TagMappingSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTagMappingSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagMappingSaveLogic {
	return &TagMappingSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagMappingSaveLogic) TagMappingSave(req *types.TagMappingSaveReq) (resp *types.BaseResp, err error) {
	doc := &model.TagMapping{
		AppName:     req.AppName,
		NucleiTags:  req.NucleiTags,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	if req.Id != "" {
		err = l.svcCtx.TagMappingModel.Update(l.ctx, req.Id, doc)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "更新失败"}, nil
		}
	} else {
		err = l.svcCtx.TagMappingModel.Insert(l.ctx, doc)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "创建失败"}, nil
		}
	}

	return &types.BaseResp{Code: 0, Msg: "保存成功"}, nil
}

type TagMappingDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTagMappingDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TagMappingDeleteLogic {
	return &TagMappingDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TagMappingDeleteLogic) TagMappingDelete(req *types.TagMappingDeleteReq) (resp *types.BaseResp, err error) {
	err = l.svcCtx.TagMappingModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// ==================== 自定义POC ====================

type CustomPocListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomPocListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomPocListLogic {
	return &CustomPocListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CustomPocListLogic) CustomPocList(req *types.CustomPocListReq) (resp *types.CustomPocListResp, err error) {
	docs, err := l.svcCtx.CustomPocModel.FindAll(l.ctx, req.Page, req.PageSize)
	if err != nil {
		return &types.CustomPocListResp{Code: 500, Msg: "查询失败"}, nil
	}

	total, _ := l.svcCtx.CustomPocModel.Count(l.ctx)

	list := make([]types.CustomPoc, 0, len(docs))
	for _, doc := range docs {
		list = append(list, types.CustomPoc{
			Id:          doc.Id.Hex(),
			Name:        doc.Name,
			TemplateId:  doc.TemplateId,
			Severity:    doc.Severity,
			Tags:        doc.Tags,
			Author:      doc.Author,
			Description: doc.Description,
			Content:     doc.Content,
			Enabled:     doc.Enabled,
			CreateTime:  doc.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.CustomPocListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

type CustomPocSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomPocSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomPocSaveLogic {
	return &CustomPocSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CustomPocSaveLogic) CustomPocSave(req *types.CustomPocSaveReq) (resp *types.BaseResp, err error) {
	doc := &model.CustomPoc{
		Name:        req.Name,
		TemplateId:  req.TemplateId,
		Severity:    req.Severity,
		Tags:        req.Tags,
		Author:      req.Author,
		Description: req.Description,
		Content:     req.Content,
		Enabled:     req.Enabled,
	}

	if req.Id != "" {
		err = l.svcCtx.CustomPocModel.Update(l.ctx, req.Id, doc)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "更新失败"}, nil
		}
	} else {
		err = l.svcCtx.CustomPocModel.Insert(l.ctx, doc)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "创建失败"}, nil
		}
	}

	return &types.BaseResp{Code: 0, Msg: "保存成功"}, nil
}

type CustomPocDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomPocDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomPocDeleteLogic {
	return &CustomPocDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CustomPocDeleteLogic) CustomPocDelete(req *types.CustomPocDeleteReq) (resp *types.BaseResp, err error) {
	err = l.svcCtx.CustomPocModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "删除失败"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "删除成功"}, nil
}

// ==================== 批量导入自定义POC ====================

type CustomPocBatchImportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomPocBatchImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomPocBatchImportLogic {
	return &CustomPocBatchImportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CustomPocBatchImportLogic) CustomPocBatchImport(req *types.CustomPocBatchImportReq) (resp *types.CustomPocBatchImportResp, err error) {
	if len(req.Pocs) == 0 {
		return &types.CustomPocBatchImportResp{Code: 400, Msg: "POC列表不能为空"}, nil
	}

	imported := 0
	failed := 0
	errors := make([]string, 0)

	for i, poc := range req.Pocs {
		// 验证必填字段
		if poc.Name == "" {
			failed++
			errors = append(errors, fmt.Sprintf("第%d个POC: 名称不能为空", i+1))
			continue
		}
		if poc.Content == "" {
			failed++
			errors = append(errors, poc.Name+": 内容不能为空")
			continue
		}

		doc := &model.CustomPoc{
			Name:        poc.Name,
			TemplateId:  poc.TemplateId,
			Severity:    poc.Severity,
			Tags:        poc.Tags,
			Author:      poc.Author,
			Description: poc.Description,
			Content:     poc.Content,
			Enabled:     poc.Enabled,
		}

		err := l.svcCtx.CustomPocModel.Insert(l.ctx, doc)
		if err != nil {
			failed++
			errors = append(errors, poc.Name+": "+err.Error())
			continue
		}
		imported++
	}

	msg := "导入完成"
	if failed > 0 {
		msg = "部分导入成功"
	}

	return &types.CustomPocBatchImportResp{
		Code:     0,
		Msg:      msg,
		Imported: imported,
		Failed:   failed,
		Errors:   errors,
	}, nil
}

// ==================== Nuclei默认模板 ====================

type NucleiTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNucleiTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NucleiTemplateListLogic {
	return &NucleiTemplateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NucleiTemplateListLogic) NucleiTemplateList(req *types.NucleiTemplateListReq) (resp *types.NucleiTemplateListResp, err error) {
	// 构建查询条件
	filter := bson.M{}
	if req.Category != "" {
		filter["category"] = req.Category
	}
	if req.Severity != "" {
		filter["severity"] = strings.ToLower(req.Severity)
	}
	if req.Tag != "" {
		// 标签模糊匹配
		filter["tags"] = bson.M{"$regex": req.Tag, "$options": "i"}
	}
	if req.Keyword != "" {
		// 使用正则表达式进行模糊搜索
		filter["$or"] = []bson.M{
			{"template_id": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"name": bson.M{"$regex": req.Keyword, "$options": "i"}},
			{"description": bson.M{"$regex": req.Keyword, "$options": "i"}},
		}
	}

	// 查询总数
	total, err := l.svcCtx.NucleiTemplateModel.Count(l.ctx, filter)
	if err != nil {
		return &types.NucleiTemplateListResp{Code: 500, Msg: "查询失败: " + err.Error()}, nil
	}

	// 查询列表
	docs, err := l.svcCtx.NucleiTemplateModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.NucleiTemplateListResp{Code: 500, Msg: "查询失败: " + err.Error()}, nil
	}

	// 转换为响应类型
	list := make([]types.NucleiTemplate, 0, len(docs))
	for _, doc := range docs {
		list = append(list, types.NucleiTemplate{
			Id:          doc.TemplateId,
			Name:        doc.Name,
			Author:      doc.Author,
			Severity:    doc.Severity,
			Description: doc.Description,
			Tags:        doc.Tags,
			Category:    doc.Category,
			FilePath:    doc.FilePath,
		})
	}

	return &types.NucleiTemplateListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

type NucleiTemplateCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNucleiTemplateCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NucleiTemplateCategoriesLogic {
	return &NucleiTemplateCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NucleiTemplateCategoriesLogic) NucleiTemplateCategories() (resp *types.NucleiTemplateCategoriesResp, err error) {
	// 使用缓存的数据
	categories := l.svcCtx.TemplateCategories
	tags := l.svcCtx.TemplateTags
	stats := l.svcCtx.TemplateStats

	if stats == nil {
		stats = map[string]int{"total": 0}
	}

	severities := []string{"critical", "high", "medium", "low", "info", "unknown"}

	return &types.NucleiTemplateCategoriesResp{
		Code:       0,
		Msg:        "success",
		Categories: categories,
		Severities: severities,
		Tags:       tags,
		Stats:      stats,
	}, nil
}


// ==================== Nuclei模板启用/禁用 ====================

type NucleiTemplateUpdateEnabledLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNucleiTemplateUpdateEnabledLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NucleiTemplateUpdateEnabledLogic {
	return &NucleiTemplateUpdateEnabledLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NucleiTemplateUpdateEnabledLogic) UpdateEnabled(req *types.NucleiTemplateUpdateEnabledReq) (resp *types.BaseResp, err error) {
	if len(req.TemplateIds) == 0 {
		return &types.BaseResp{Code: 400, Msg: "请选择模板"}, nil
	}

	err = l.svcCtx.NucleiTemplateModel.BatchUpdateEnabled(l.ctx, req.TemplateIds, req.Enabled)
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "更新失败: " + err.Error()}, nil
	}

	action := "启用"
	if !req.Enabled {
		action = "禁用"
	}
	return &types.BaseResp{Code: 0, Msg: action + "成功"}, nil
}


// ==================== Nuclei模板详情 ====================

type NucleiTemplateDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNucleiTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NucleiTemplateDetailLogic {
	return &NucleiTemplateDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NucleiTemplateDetailLogic) GetDetail(req *types.NucleiTemplateDetailReq) (resp *types.NucleiTemplateDetailResp, err error) {
	if req.TemplateId == "" {
		return &types.NucleiTemplateDetailResp{Code: 400, Msg: "模板ID不能为空"}, nil
	}

	// 从数据库查询完整模板（包含content）
	doc, err := l.svcCtx.NucleiTemplateModel.FindByTemplateId(l.ctx, req.TemplateId)
	if err != nil || doc == nil {
		return &types.NucleiTemplateDetailResp{Code: 404, Msg: "模板不存在"}, nil
	}
	return &types.NucleiTemplateDetailResp{
		Code: 0,
		Msg:  "success",
		Data: &types.NucleiTemplateWithContent{
			Id:          doc.TemplateId,
			Name:        doc.Name,
			Author:      doc.Author,
			Severity:    doc.Severity,
			Description: doc.Description,
			Tags:        doc.Tags,
			FilePath:    doc.FilePath,
			Content:     doc.Content,
		},
	}, nil
}

// ==================== POC验证 ====================

type PocValidateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPocValidateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PocValidateLogic {
	return &PocValidateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PocValidateLogic) PocValidate(req *types.PocValidateReq) (resp *types.PocValidateResp, err error) {
	if req.Url == "" {
		return &types.PocValidateResp{Code: 400, Msg: "URL不能为空"}, nil
	}
	if req.Id == "" {
		return &types.PocValidateResp{Code: 400, Msg: "POC ID不能为空"}, nil
	}

	// 根据pocType确定POC类型
	pocType := req.PocType
	if pocType == "" {
		pocType = "custom" // 默认为自定义POC
	}

	var pocSeverity string

	if pocType == "nuclei" {
		// Nuclei默认模板
		template, err := l.svcCtx.NucleiTemplateModel.FindByTemplateId(l.ctx, req.Id)
		if err != nil {
			return &types.PocValidateResp{Code: 404, Msg: "Nuclei模板不存在"}, nil
		}
		pocSeverity = template.Severity
	} else {
		// 自定义POC
		poc, err := l.svcCtx.CustomPocModel.FindById(l.ctx, req.Id)
		if err != nil {
			return &types.PocValidateResp{Code: 404, Msg: "POC不存在"}, nil
		}
		pocSeverity = poc.Severity
	}

	// 通过RPC调用worker执行POC验证
	rpcReq := &pb.ValidatePocReq{
		Url:         req.Url,
		PocId:       req.Id,
		PocType:     pocType,
		Timeout:     30,
		UseTemplate: pocType == "nuclei",
		UseCustom:   pocType == "custom",
	}

	rpcResp, err := l.svcCtx.TaskRpcClient.ValidatePoc(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("RPC call failed: %v", err)
		return &types.PocValidateResp{Code: 500, Msg: "验证服务调用失败"}, nil
	}

	if !rpcResp.Success {
		return &types.PocValidateResp{Code: 500, Msg: rpcResp.Message}, nil
	}

	// 异步模式：返回任务已下发的信息和任务ID
	return &types.PocValidateResp{
		Code:     0,
		Msg:      "POC验证任务已下发，请稍后查询结果",
		Matched:  false, // 异步模式下无法立即返回匹配结果
		Severity: pocSeverity,
		Details:  rpcResp.Details,
		TaskId:   rpcResp.TaskId,
	}, nil
}
// ==================== 批量POC验证 ====================

type PocBatchValidateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPocBatchValidateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PocBatchValidateLogic {
	return &PocBatchValidateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PocBatchValidateLogic) PocBatchValidate(req *types.PocBatchValidateReq) (resp *types.PocBatchValidateResp, err error) {
	if len(req.Urls) == 0 {
		return &types.PocBatchValidateResp{Code: 400, Msg: "URL列表不能为空"}, nil
	}

	// 设置默认值
	if req.PocType == "" {
		req.PocType = "all"
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 10
	}
	if req.UseTemplate == false && req.UseCustom == false {
		req.UseTemplate = true
		req.UseCustom = true
	}

	// 通过RPC调用worker执行批量POC验证
	rpcReq := &pb.BatchValidatePocReq{
		Urls:        req.Urls,
		PocType:     req.PocType,
		Severities:  req.Severities,
		Tags:        req.Tags,
		Timeout:     int32(req.Timeout),
		UseTemplate: req.UseTemplate,
		UseCustom:   req.UseCustom,
		Concurrency: int32(req.Concurrency),
	}

	rpcResp, err := l.svcCtx.TaskRpcClient.BatchValidatePoc(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("RPC call failed: %v", err)
		return &types.PocBatchValidateResp{Code: 500, Msg: "验证服务调用失败"}, nil
	}

	if !rpcResp.Success {
		return &types.PocBatchValidateResp{Code: 500, Msg: rpcResp.Message}, nil
	}

	// 从RPC响应中获取批次ID
	batchId := rpcResp.BatchId

	return &types.PocBatchValidateResp{
		Code:      0,
		Msg:       "批量验证任务已下发，请使用返回的批次ID查询结果",
		TotalUrls: int(rpcResp.TotalUrls),
		Duration:  rpcResp.Duration,
		BatchId:   batchId,
	}, nil
}
// ==================== POC验证结果查询 ====================

type PocValidationResultQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPocValidationResultQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PocValidationResultQueryLogic {
	return &PocValidationResultQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PocValidationResultQueryLogic) PocValidationResultQuery(req *types.PocValidationResultQueryReq) (resp *types.PocValidationResultQueryResp, err error) {
	if req.TaskId == "" && req.BatchId == "" {
		return &types.PocValidationResultQueryResp{Code: 400, Msg: "任务ID或批次ID不能为空"}, nil
	}

	// 通过RPC查询验证结果
	rpcReq := &pb.GetPocValidationResultReq{
		TaskId:  req.TaskId,
		BatchId: req.BatchId,
	}

	rpcResp, err := l.svcCtx.TaskRpcClient.GetPocValidationResult(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("RPC call failed: %v", err)
		return &types.PocValidationResultQueryResp{Code: 500, Msg: "查询服务调用失败"}, nil
	}

	if !rpcResp.Success {
		return &types.PocValidationResultQueryResp{Code: 500, Msg: rpcResp.Message}, nil
	}

	// 转换结果
	results := make([]types.PocValidationResult, 0, len(rpcResp.Results))
	for _, r := range rpcResp.Results {
		results = append(results, types.PocValidationResult{
			PocId:      r.PocId,
			PocName:    r.PocName,
			TemplateId: r.TemplateId,
			Severity:   r.Severity,
			Matched:    r.Matched,
			MatchedUrl: r.MatchedUrl,
			Details:    r.Details,
			Output:     r.Output,
			PocType:    r.PocType,
			Tags:       r.Tags,
		})
	}

	return &types.PocValidationResultQueryResp{
		Code:           0,
		Msg:            "查询成功",
		Status:         rpcResp.Status,
		CompletedCount: int(rpcResp.CompletedCount),
		TotalCount:     int(rpcResp.TotalCount),
		Results:        results,
		CreateTime:     rpcResp.CreateTime,
		UpdateTime:     rpcResp.UpdateTime,
	}, nil
}


// ==================== 清空所有自定义POC ====================

type CustomPocClearAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomPocClearAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomPocClearAllLogic {
	return &CustomPocClearAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CustomPocClearAllLogic) CustomPocClearAll() (resp *types.CustomPocClearAllResp, err error) {
	// 先获取总数
	total, _ := l.svcCtx.CustomPocModel.Count(l.ctx)
	
	// 删除所有自定义POC
	deleted, err := l.svcCtx.CustomPocModel.DeleteAll(l.ctx)
	if err != nil {
		return &types.CustomPocClearAllResp{Code: 500, Msg: "清空失败: " + err.Error()}, nil
	}
	
	if deleted == 0 {
		deleted = total
	}
	
	return &types.CustomPocClearAllResp{
		Code:    0,
		Msg:     "清空成功",
		Deleted: int(deleted),
	}, nil
}
