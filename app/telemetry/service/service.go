package service

import (
	"context"
	"fmt"

	"github.com/flussrd/fluss-back/app/river-management/handlers/grpc/grpchandler"
	riverGrpcClient "github.com/flussrd/fluss-back/app/shared/grpc-clients/river-management"
	"github.com/flussrd/fluss-back/app/telemetry/models"
	repository "github.com/flussrd/fluss-back/app/telemetry/repositories/measurements"
)

type service struct {
	riverClient *riverGrpcClient.Client
	repo        repository.Repository
}

func New(riverClient *riverGrpcClient.Client, repo repository.Repository) Service {
	return service{
		riverClient: riverClient,
		repo:        repo,
	}
}

func (s service) SaveMeasurement(ctx context.Context, message models.Message) error {
	// TODO: validate message before calling these functions
	client := s.riverClient.GetServiceClient()
	module, err := client.GetModuleByPhonenumber(ctx, &grpchandler.GetModuleRequest{PhoneNumber: message.PhoneNumber})
	if err != nil {
		fmt.Println("getting_module_by_phone_number_failed: ", err.Error())
		return err
	}

	message.ModuleID = module.ModuleID

	return s.repo.SaveMeasurement(ctx, message)
}