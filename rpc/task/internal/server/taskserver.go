package server

import (
	"context"

	"cscan/rpc/task/internal/logic"
	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"
)

type TaskServiceServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedTaskServiceServer
}

func NewTaskServiceServer(svcCtx *svc.ServiceContext) *TaskServiceServer {
	return &TaskServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *TaskServiceServer) CheckTask(ctx context.Context, in *pb.CheckTaskReq) (*pb.CheckTaskResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.CheckTask(in)
}

func (s *TaskServiceServer) UpdateTask(ctx context.Context, in *pb.UpdateTaskReq) (*pb.UpdateTaskResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.UpdateTask(in)
}

func (s *TaskServiceServer) NewTask(ctx context.Context, in *pb.NewTaskReq) (*pb.NewTaskResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.NewTask(in)
}

func (s *TaskServiceServer) SaveTaskResult(ctx context.Context, in *pb.SaveTaskResultReq) (*pb.SaveTaskResultResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.SaveTaskResult(in)
}

func (s *TaskServiceServer) SaveVulResult(ctx context.Context, in *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.SaveVulResult(in)
}

func (s *TaskServiceServer) KeepAlive(ctx context.Context, in *pb.KeepAliveReq) (*pb.KeepAliveResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.KeepAlive(in)
}

func (s *TaskServiceServer) GetWorkerConfig(ctx context.Context, in *pb.GetWorkerConfigReq) (*pb.GetWorkerConfigResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetWorkerConfig(in)
}

func (s *TaskServiceServer) RequestResource(ctx context.Context, in *pb.RequestResourceReq) (*pb.RequestResourceResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.RequestResource(in)
}

func (s *TaskServiceServer) GetTemplatesByTags(ctx context.Context, in *pb.GetTemplatesByTagsReq) (*pb.GetTemplatesByTagsResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetTemplatesByTags(in)
}

func (s *TaskServiceServer) GetCustomFingerprints(ctx context.Context, in *pb.GetCustomFingerprintsReq) (*pb.GetCustomFingerprintsResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetCustomFingerprints(in)
}

func (s *TaskServiceServer) ValidateFingerprint(ctx context.Context, in *pb.ValidateFingerprintReq) (*pb.ValidateFingerprintResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.ValidateFingerprint(in)
}
func (s *TaskServiceServer) ValidatePoc(ctx context.Context, in *pb.ValidatePocReq) (*pb.ValidatePocResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.ValidatePoc(in)
}

func (s *TaskServiceServer) BatchValidatePoc(ctx context.Context, in *pb.BatchValidatePocReq) (*pb.BatchValidatePocResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.BatchValidatePoc(in)
}
func (s *TaskServiceServer) GetPocValidationResult(ctx context.Context, in *pb.GetPocValidationResultReq) (*pb.GetPocValidationResultResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetPocValidationResult(in)
}


func (s *TaskServiceServer) GetPocById(ctx context.Context, in *pb.GetPocByIdReq) (*pb.GetPocByIdResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetPocById(in)
}

func (s *TaskServiceServer) GetTemplatesByIds(ctx context.Context, in *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetTemplatesByIds(in)
}

func (s *TaskServiceServer) GetHttpServiceMappings(ctx context.Context, in *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error) {
	l := logic.NewTaskLogic(ctx, s.svcCtx)
	return l.GetHttpServiceMappings(in)
}
