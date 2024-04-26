//go:generate mockgen -destination=grpc_mocks_test.go -package=grpc github.com/TutorialEdge/go-grpc-services-course/internal/transport/grpc RocketService

package grpc

import (
	"context"
	"log"
	"net"

	rkt "github.com/TutorialEdge/tutorial-protos/rocket/v1"
	"github.com/google/uuid"
	"github.com/moabukar/grpc/internal/rocket"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RocketService - define the interface that the concrete implementation
// has to adhere to
type RocketService interface {
	GetRocketByID(ctx context.Context, id string) (rocket.Rocket, error)
	InsertRocket(ctx context.Context, rkt rocket.Rocket) (rocket.Rocket, error)
	DeleteRocket(ctx context.Context, id string) error
}

// Handler - will handle incoming gRPC requests
type Handler struct {
	RocketService RocketService
}

// New - returns a new gRPC handler
func New(rktService RocketService) Handler {
	return Handler{
		RocketService: rktService,
	}
}

func (h Handler) Serve() error {
	address := ":50051" // Ensure this is dynamically set if necessary
	log.Printf("Attempting to listen on %s", address)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Print("could not listen on port 50051")
		return err
	}

	grpcServer := grpc.NewServer()
	rkt.RegisterRocketServiceServer(grpcServer, &h)

	if err := grpcServer.Serve(lis); err != nil {
		log.Printf("failed to serve: %s\n", err)
		return err
	}

	return nil
}

// GetRocket - retrieves a rocket by id and returns the response.
func (h Handler) GetRocket(ctx context.Context, req *rkt.GetRocketRequest) (*rkt.GetRocketResponse, error) {
	log.Print("Get Rocket gRPC Endpoint Hit")

	_, err := uuid.Parse(req.Id)
	if err != nil {
		log.Printf("Given UUID is not valid: %v", err)
		errorStatus := status.New(codes.InvalidArgument, "UUID is not valid")
		details, err := errorStatus.WithDetails(&errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "id",
					Description: "UUID is not valid",
				},
			},
		})
		if err != nil {
			log.Printf("Error adding details to error status: %v", err)
			return &rkt.GetRocketResponse{}, err
		}
		return &rkt.GetRocketResponse{}, details.Err()
	}

	rocket, err := h.RocketService.GetRocketByID(ctx, req.Id)
	if err != nil {
		log.Printf("Failed to retrieve rocket by ID: %v", err)
		return &rkt.GetRocketResponse{}, status.Error(codes.Internal, "internal server error")
	}

	return &rkt.GetRocketResponse{
		Rocket: &rkt.Rocket{
			Id:   rocket.ID,
			Name: rocket.Name,
			Type: rocket.Type,
		},
	}, nil
}

// AddRocket - adds a rocket to the database
func (h Handler) AddRocket(ctx context.Context, req *rkt.AddRocketRequest) (*rkt.AddRocketResponse, error) {
	log.Print("Add Rocket gRPC endpoint hit")

	if _, err := uuid.Parse(req.Rocket.Id); err != nil {
		errorStatus := status.Error(codes.InvalidArgument, "UUID is not valid")
		log.Print("Given UUID is not valid")
		return &rkt.AddRocketResponse{}, errorStatus
	}

	newRkt, err := h.RocketService.InsertRocket(ctx, rocket.Rocket{
		ID:   req.Rocket.Id,
		Type: req.Rocket.Type,
		Name: req.Rocket.Name,
	})
	if err != nil {
		log.Print("failed to insert rocket into database")
		return &rkt.AddRocketResponse{}, err
	}
	return &rkt.AddRocketResponse{
		Rocket: &rkt.Rocket{
			Id:   newRkt.ID,
			Type: newRkt.Type,
			Name: newRkt.Name,
		},
	}, nil
}

// DeleteRocket - handler for deleting a rocket
func (h Handler) DeleteRocket(ctx context.Context, req *rkt.DeleteRocketRequest) (*rkt.DeleteRocketResponse, error) {
	log.Print("delete rocket gRPC endpoint hit")
	err := h.RocketService.DeleteRocket(ctx, req.Rocket.Id)
	if err != nil {
		return &rkt.DeleteRocketResponse{}, err
	}
	return &rkt.DeleteRocketResponse{
		Status: "successfully delete rocket",
	}, nil
}
