package CosmosWallet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type CosmosWallet struct {
	address          string
	validatorAddress string
}

func NewWallet() *CosmosWallet {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// wallet (Delegator) - 4935033 ATOM
	address := os.Getenv("DELEGATOR")
	// Validator (staking)
	validatorAddress := os.Getenv("VALIDATOR")

	fmt.Printf("Wallet address: %v\n", address)
	fmt.Printf("Validator address: %v\n", validatorAddress)

	return &CosmosWallet{
		address: address,

		validatorAddress: validatorAddress,
	}

}

func (wallet CosmosWallet) GetAvaliableTokens(grpcConn *grpc.ClientConn) (int, error) {
	client := banktypes.NewQueryClient(grpcConn)
	grpcRes, err := client.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{
			Address: wallet.address,
			Denom:   "uatom",
		},
	)
	if err != nil {
		fmt.Printf("Get balance error: %v", err)
	}
	return int(grpcRes.Balance.Amount.Uint64()), nil

}

func (wallet CosmosWallet) ClaimStakingRewards(grpcConn *grpc.ClientConn) (int, error) {
	fmt.Printf("Claiming staking rewards\n")

	client := distributiontypes.NewQueryClient(grpcConn)
	grpcRes, err := client.DelegationRewards(
		context.Background(),
		&distributiontypes.QueryDelegationRewardsRequest{
			DelegatorAddress: wallet.address,
			ValidatorAddress: wallet.validatorAddress,
		},
	)
	if err != nil {
		return 0, err
	}

	coins, _ := grpcRes.Rewards.TruncateDecimal()
	reward := int(coins.AmountOf("uatom").Uint64())

	// I get this value from the transaction that I made to claim the rewards. I shoud get this programatically.
	// gasCostEstimation := 504072
	gasCostEstimation := 1

	// 2 times, becousa i claims the rewards and restake the rewards. If the value is less than 2 times, it will be impossible to restake the rewards or the rewards will be lost, because the gas cost is higher than the rewards.
	if reward <= gasCostEstimation*2 {
		return 0, errors.New("not enough rewards to claim")
	}

	// {
	// 	"account_number": "2695013",
	// 	"chain_id": "cosmoshub-4",
	// 	"fee": {
	// 	  "gas": "504072",
	// 	  "amount": [
	// 		{
	// 		  "amount": "2521",
	// 		  "denom": "uatom"
	// 		}
	// 	  ]
	// 	},
	// 	"memo": "",
	// 	"msgs": [
	// 	  {
	// 		"type": "cosmos-sdk/MsgWithdrawDelegationReward",
	// 		"value": {
	// 		  "delegator_address": "cosmos1gqu2kthx340yu7gnjhnj4vfuxk50kuy90mf76a",
	// 		  "validator_address": "cosmosvaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3g8fgh9"
	// 		}
	// 	  }
	// 	],
	// 	"sequence": "3"
	//   }

	return reward, nil
}

func (wallet CosmosWallet) Stake(grpcConn *grpc.ClientConn, tokens int) error {
	// validator, err := wallet.getStakeValidator(grpcConn)
	_, err := wallet.getStakeValidator(grpcConn)
	if err != nil {
		fmt.Printf("Error getting stake validator: %s\n", err)
		return err
	}
	// fmt.Printf("Validator: %v\n", validator)

	staked, err := wallet.getStakedAmount(grpcConn)
	if err != nil {
		fmt.Printf("Error when try to get staked value: %v\n", err)
		return nil
	}
	fmt.Printf("Staked amount: %v\n", staked)

	// It is not possible stake using gPRC. I need to create a transaction, sing then use gRCP to broadcast the transaction.
	// https://docs.cosmos.network/v0.46/run-node/txs.html#using-grpc

	// Delegate msg
	// {
	// 	"account_number": "2695013",
	// 	"chain_id": "cosmoshub-4",
	// 	"fee": {
	// 	  "gas": "581186",
	// 	  "amount": [
	// 		{
	// 		  "amount": "14530",
	// 		  "denom": "uatom"
	// 		}
	// 	  ]
	// 	},
	// 	"memo": "",
	// 	"msgs": [
	// 	  {
	// 		"type": "cosmos-sdk/MsgDelegate",
	// 		"value": {
	// 		  "amount": {
	// 			"amount": "5000",
	// 			"denom": "uatom"
	// 		  },
	// 		  "delegator_address": "cosmos1gqu2kthx340yu7gnjhnj4vfuxk50kuy90mf76a",
	// 		  "validator_address": "cosmosvaloper1de7qx00pz2j6gn9k88ntxxylelkazfk3g8fgh9"
	// 		}
	// 	  }
	// 	],
	// 	"sequence": "3"
	//   }

	return nil
}

func (wallet CosmosWallet) getStakeValidator(grpcConn *grpc.ClientConn) (*stakingtypes.Validator, error) {
	client := stakingtypes.NewQueryClient(grpcConn)
	grpcRes, err := client.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{
			ValidatorAddr: wallet.validatorAddress,
		},
	)
	if err != nil {
		fmt.Printf("Get balance error: %v", err)
	}
	return &grpcRes.Validator, nil
}

func (wallet CosmosWallet) getStakedAmount(grpcConn *grpc.ClientConn) (int, error) {
	client := stakingtypes.NewQueryClient(grpcConn)
	grpcRes, err := client.Delegation(
		context.Background(),
		&stakingtypes.QueryDelegationRequest{
			DelegatorAddr: wallet.address,
			ValidatorAddr: wallet.validatorAddress,
		},
	)
	if err != nil {
		return 0, err

	}
	return int(grpcRes.DelegationResponse.Balance.Amount.Uint64()), nil
}

func (wallet CosmosWallet) transfer(grpcConn *grpc.ClientConn) error {
	from := sdk.AccAddress("cosmos1gqu2kthx340yu7gnjhnj4vfuxk50kuy90mf76a")
	to := sdk.AccAddress("cosmos13z4qcmn2hmvgn0p50p7a20dcqwaefj9k2zn8fq")
	amount := sdk.NewCoins(sdk.NewInt64Coin("uatom", 100))
	client := banktypes.NewMsgSend(from, to, amount)

	fmt.Printf("transfer Client: %v\n", client)
	// c := txypes.DefaultTxDecoder()
	// txBuilder := txypes.NewTxConfig(c)

	// err := txBuilder.SetMsgs(msg1, msg2)
	// if err != nil {
	//     return err
	// }

	// Send msg
	// {
	// 	"account_number": "2695013",
	// 	"chain_id": "cosmoshub-4",
	// 	"fee": {
	// 	  "gas": "70069",
	// 	  "amount": [
	// 		{
	// 		  "amount": "1752",
	// 		  "denom": "uatom"
	// 		}
	// 	  ]
	// 	},
	// 	"memo": "",
	// 	"msgs": [
	// 	  {
	// 		"type": "cosmos-sdk/MsgSend",
	// 		"value": {
	// 		  "amount": [
	// 			{
	// 			  "amount": "100",
	// 			  "denom": "uatom"
	// 			}
	// 		  ],
	// 		  "from_address": "cosmos1gqu2kthx340yu7gnjhnj4vfuxk50kuy90mf76a",
	// 		  "to_address": "cosmos13z4qcmn2hmvgn0p50p7a20dcqwaefj9k2zn8fq"
	// 		}
	// 	  }
	// 	],
	// 	"sequence": "4"
	//   }

	return nil
}

// encript and sign transaction
func (wallet CosmosWallet) sign() error {
	fmt.Print("sing transaction\n")
	// My intention is to use the Cosmos SDK to sign the transaction.
	// I find same examples how to sing a transaction using the Cosmos SDK, but tey are not clear.
	// this one is the closest (https://docs.cosmos.network/main/user/run-node/txs#programmatically-with-go) but it is not clear/incomplete.
	// On this example, the transaction is signed the "txBuilder" whitch is created by the "simapp". In the end, I couldn't make it work.
	return nil
}

func (wallet CosmosWallet) broadcast(grpcConn *grpc.ClientConn) error {
	fmt.Print("Broadcasting transaction\n")

	// txtypes.

	// client := txtypes.NewQueryClient(grpcConn)
	// grpcRes, err := client.Validator(
	// 	context.Background(),
	// 	&stakingtypes.QueryValidatorRequest{
	// 		ValidatorAddr: wallet.validatorAddress,
	// 	},
	// )
	// if err != nil {
	// 	fmt.Printf("Get balance error: %v", err)
	// }

	return nil
}
