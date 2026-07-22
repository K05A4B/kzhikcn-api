package service

type TopicService struct {
	ctx *ServiceContext
}

func NewTopicService(ctx *ServiceContext) *TopicService {
	return &TopicService{ctx: ctx}
}
