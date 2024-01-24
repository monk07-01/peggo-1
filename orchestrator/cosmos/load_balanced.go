package cosmos

import (
	"context"
	"github.com/InjectiveLabs/peggo/orchestrator/cosmos/peggy"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/InjectiveLabs/peggo/orchestrator/ethereum/keystore"
	"github.com/InjectiveLabs/sdk-go/chain/peggy/types"
	peggytypes "github.com/InjectiveLabs/sdk-go/chain/peggy/types"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/InjectiveLabs/sdk-go/client/common"
	explorerclient "github.com/InjectiveLabs/sdk-go/client/explorer"
)

type LoadBalancedNetwork struct {
	addr sdk.AccAddress

	PeggyQueryClient
	PeggyBroadcastClient
	explorerclient.ExplorerClient
}

// NewLoadBalancedNetwork creates a load balanced connection to the Injective network.
// The chainID argument decides which network Peggo will be connecting to:
//   - injective-1 (mainnet)
//   - injective-777 (devnet)
//   - injective-888 (testnet)
func NewLoadBalancedNetwork(
	chainID,
	validatorAddress,
	injectiveGasPrices string,
	keyring keyring.Keyring,
	personalSignerFn keystore.PersonalSignFn,
) (*LoadBalancedNetwork, error) {
	addr, err := sdk.AccAddressFromBech32(validatorAddress)
	if err != nil {
		return nil, errors.Wrap(err, "invalid address")
	}

	var networkName string
	switch chainID {
	case "injective-1":
		networkName = "mainnet"
	case "injective-777":
		networkName = "devnet"
	case "injective-888":
		networkName = "testnet"
	default:
		return nil, errors.Errorf("provided chain id %v does not belong to any known Injective network", chainID)
	}

	netCfg := common.LoadNetwork(networkName, "lb")
	explorer, err := explorerclient.NewExplorerClient(netCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize explorer client")
	}

	clientCtx, err := chainclient.NewClientContext(chainID, validatorAddress, keyring)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create client context for Injective chain")
	}

	tmClient, err := rpchttp.New(netCfg.TmEndpoint, "/websocket")
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize tendermint client")
	}

	clientCtx = clientCtx.WithNodeURI(netCfg.TmEndpoint).WithClient(tmClient)

	daemonClient, err := chainclient.NewChainClient(clientCtx, netCfg, common.OptionGasPrices(injectiveGasPrices))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to intialize chain client (%s)", networkName)
	}

	time.Sleep(1 * time.Second)

	daemonWaitCtx, cancelWait := context.WithTimeout(context.Background(), time.Minute)
	defer cancelWait()

	grpcConn := daemonClient.QueryClient()
	waitForService(daemonWaitCtx, grpcConn)
	peggyQuerier := types.NewQueryClient(grpcConn)

	n := &LoadBalancedNetwork{
		addr:                 addr,
		PeggyQueryClient:     peggy.NewQueryClient(peggyQuerier),
		PeggyBroadcastClient: peggy.NewBroadcastClient(daemonClient, personalSignerFn),
		ExplorerClient:       explorer,
	}

	log.WithFields(log.Fields{
		"addr":       validatorAddress,
		"chain_id":   chainID,
		"injective":  netCfg.ChainGrpcEndpoint,
		"tendermint": netCfg.TmEndpoint,
	}).Infoln("connected to Injective's load balanced endpoints")

	return n, nil
}

func (n *LoadBalancedNetwork) GetBlockCreationTime(ctx context.Context, height int64) (time.Time, error) {
	block, err := n.ExplorerClient.GetBlock(ctx, strconv.FormatInt(height, 10))
	if err != nil {
		return time.Time{}, err
	}

	blockTime, err := time.Parse("2006-01-02 15:04:05.999 -0700 MST", block.Data.Timestamp)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse timestamp from block")
	}

	return blockTime, nil
}

func (n *LoadBalancedNetwork) LastClaimEvent(ctx context.Context) (*peggytypes.LastClaimEvent, error) {
	return n.LastClaimEventByAddr(ctx, n.addr)
}

func (n *LoadBalancedNetwork) OldestUnsignedValsets(ctx context.Context) ([]*peggytypes.Valset, error) {
	return n.PeggyQueryClient.OldestUnsignedValsets(ctx, n.addr)
}

func (n *LoadBalancedNetwork) OldestUnsignedTransactionBatch(ctx context.Context) (*peggytypes.OutgoingTxBatch, error) {
	return n.PeggyQueryClient.OldestUnsignedTransactionBatch(ctx, n.addr)
}

// waitForService awaits an active ClientConn to a GRPC service.
func waitForService(ctx context.Context, clientConn *grpc.ClientConn) {
	for {
		select {
		case <-ctx.Done():
			log.Fatalln("GRPC service wait timed out")
		default:
			state := clientConn.GetState()

			if state != connectivity.Ready {
				log.WithField("state", state.String()).Warningln("state of GRPC connection not ready")
				time.Sleep(5 * time.Second)
				continue
			}

			return
		}
	}
}
