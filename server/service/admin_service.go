package service

type AdminService struct {
	ctx *ServiceContext
}

func NewAdminService(ctx *ServiceContext) *AdminService {
	return &AdminService{ctx: ctx}
}
