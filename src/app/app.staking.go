package app

import (
	"fmt"

	"github.com/yardz/test-apybara-go-bot/src/modules/CosmosWallet"
	"google.golang.org/grpc"
)

type App struct {
	wallet *CosmosWallet.CosmosWallet
}

func NewApp() *App {

	return &App{
		wallet: CosmosWallet.NewWallet(),
	}
}

func (app *App) RunStakeBot() {
	grpcConn := app.GetGRPCconnection()
	defer grpcConn.Close()

	rewards, err := app.wallet.ClaimStakingRewards(grpcConn)
	if err != nil {
		fmt.Printf("Error claiming staking rewards: %s\n", err)
		return
	}
	fmt.Printf("Claimed staking rewards: %d\n", rewards)

	avaliableTokens, err := app.wallet.GetAvaliableTokens(grpcConn)
	if err != nil {
		fmt.Printf("Error getting available tokens: %s\n", err)
	}
	if avaliableTokens == 0 {
		fmt.Printf("No tokens available\n")
		return
	}
	fmt.Printf("Available tokens: %d\n", avaliableTokens)

	// This is just a validation to avoid staking more tokens than the rewards.
	if avaliableTokens < rewards {
		fmt.Printf("Not enough tokens to restake\n")
		return
	}

	err = app.wallet.Stake(grpcConn, rewards)
	if err != nil {
		fmt.Printf("Error staking tokens: %s\n", err)
	}
}

func (app *App) GetGRPCconnection() *grpc.ClientConn {
	grpcConn, err := grpc.Dial(
		// "cosmos-grpc.stakeandrelax.net:14990",
		"cosmos-grpc.polkachu.com:14990",
		grpc.WithInsecure(),
	)
	if err != nil {
		fmt.Printf("GRPC connection error: %v", err)
	}

	return grpcConn
}
