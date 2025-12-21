package client

import (
	"context"

	"cscan/rpc/task/pb"
)

// TaskClient 任务客户端接口
type TaskClient interface {
	// CheckTask 检查任务
	CheckTask(ctx context.Context, req *pb.CheckTaskReq) (*pb.CheckTaskResp, error)
	// UpdateTask 更新任务状态
	UpdateTask(ctx context.Context, req *pb.UpdateTaskReq) (*pb.UpdateTaskResp, error)
	// SaveTaskResult 保存任务结果
	SaveTaskResult(ctx context.Context, req *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error)
	// SaveVulResult 保存漏洞结果
	SaveVulResult(ctx context.Context, req *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error)
	// KeepAlive 心跳
	KeepAlive(ctx context.Context, req *pb.KeepAliveReq) (*pb.KeepAliveResp, error)
	// GetTemplatesByTags 根据标签获取模板
	GetTemplatesByTags(ctx context.Context, req *pb.GetTemplatesByTagsReq) (*pb.GetTemplatesByTagsResp, error)
	// GetTemplatesByIds 根据ID获取模板
	GetTemplatesByIds(ctx context.Context, req *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error)
	// GetCustomFingerprints 获取自定义指纹
	GetCustomFingerprints(ctx context.Context, req *pb.GetCustomFingerprintsReq) (*pb.GetCustomFingerprintsResp, error)
	// GetHttpServiceMappings 获取HTTP服务映射
	GetHttpServiceMappings(ctx context.Context, req *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error)
}

// RpcTaskClient RPC实现的任务客户端
type RpcTaskClient struct {
	client pb.TaskServiceClient
}

// NewRpcTaskClient 创建RPC任务客户端
func NewRpcTaskClient(client pb.TaskServiceClient) TaskClient {
	return &RpcTaskClient{client: client}
}

func (c *RpcTaskClient) CheckTask(ctx context.Context, req *pb.CheckTaskReq) (*pb.CheckTaskResp, error) {
	return c.client.CheckTask(ctx, req)
}

func (c *RpcTaskClient) UpdateTask(ctx context.Context, req *pb.UpdateTaskReq) (*pb.UpdateTaskResp, error) {
	return c.client.UpdateTask(ctx, req)
}

func (c *RpcTaskClient) SaveTaskResult(ctx context.Context, req *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error) {
	return c.client.SaveTaskResult(ctx, req)
}

func (c *RpcTaskClient) SaveVulResult(ctx context.Context, req *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error) {
	return c.client.SaveVulResult(ctx, req)
}

func (c *RpcTaskClient) KeepAlive(ctx context.Context, req *pb.KeepAliveReq) (*pb.KeepAliveResp, error) {
	return c.client.KeepAlive(ctx, req)
}

func (c *RpcTaskClient) GetTemplatesByTags(ctx context.Context, req *pb.GetTemplatesByTagsReq) (*pb.GetTemplatesByTagsResp, error) {
	return c.client.GetTemplatesByTags(ctx, req)
}

func (c *RpcTaskClient) GetTemplatesByIds(ctx context.Context, req *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error) {
	return c.client.GetTemplatesByIds(ctx, req)
}

func (c *RpcTaskClient) GetCustomFingerprints(ctx context.Context, req *pb.GetCustomFingerprintsReq) (*pb.GetCustomFingerprintsResp, error) {
	return c.client.GetCustomFingerprints(ctx, req)
}

func (c *RpcTaskClient) GetHttpServiceMappings(ctx context.Context, req *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error) {
	return c.client.GetHttpServiceMappings(ctx, req)
}
