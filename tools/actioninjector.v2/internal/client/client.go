package client

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/pkg/log"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

// Client is the blockchain API client.
type Client struct {
	api iotexapi.APIServiceClient
}

// New creates a new Client.
func New(serverAddr string, insecure bool) (*Client, error) {
	grpcctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error
	log.L().Info("Server Addr", zap.String("endpoint", serverAddr))
	if insecure {
		log.L().Info("insecure connection")
		conn, err = grpc.DialContext(grpcctx, serverAddr, grpc.WithBlock(), grpc.WithInsecure())
	} else {
		log.L().Info("secure connection")
		conn, err = grpc.DialContext(grpcctx, serverAddr, grpc.WithBlock(), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	if err != nil {
		return nil, err
	}
	log.L().Info("server connected")
	return &Client{
		api: iotexapi.NewAPIServiceClient(conn),
	}, nil
}

// SendAction sends an action to blockchain.
func (c *Client) SendAction(ctx context.Context, selp action.SealedEnvelope) error {
	_, err := c.api.SendAction(ctx, &iotexapi.SendActionRequest{Action: selp.Proto()})
	return err
}

// GetAccount returns a given account.
func (c *Client) GetAccount(ctx context.Context, addr string) (*iotexapi.GetAccountResponse, error) {
	return c.api.GetAccount(ctx, &iotexapi.GetAccountRequest{Address: addr})
}

// GetGasPrice returns a given account.
func (c *Client) GetGasPrice() (uint64, error) {
	res, err := c.api.SuggestGasPrice(context.Background(), &iotexapi.SuggestGasPriceRequest{})
	if err != nil {
		return 0, err
	}
	return res.GasPrice, nil
}

// GetGasPrice returns a given account.
func (c *Client) EstimateGas(caller string, act action.Action) (uint64, error) {
	var (
		ret = &iotexapi.EstimateActionGasConsumptionResponse{}
		err error
	)
	switch tx := act.(type) {
	case *action.Execution:
		log.L().Info("estimate..")
		ret, err = c.api.EstimateActionGasConsumption(context.Background(), &iotexapi.EstimateActionGasConsumptionRequest{
			Action: &iotexapi.EstimateActionGasConsumptionRequest_Execution{
				Execution: tx.Proto(),
			},
			CallerAddress: caller,
		})
	case *action.Transfer:
		ret, err = c.api.EstimateActionGasConsumption(context.Background(), &iotexapi.EstimateActionGasConsumptionRequest{
			Action: &iotexapi.EstimateActionGasConsumptionRequest_Transfer{
				Transfer: tx.Proto(),
			},
			CallerAddress: caller,
		})
	default:
		return 0, errors.New("unsupported")
	}
	return ret.GetGas(), err
}
