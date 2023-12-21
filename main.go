package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RPCURL      string          `yaml:"rpcURL"`
	PrivateKey  string          `yaml:"privateKey"`
	TransferTo  string          `yaml:"transferTo"`
	Inscription string          `yaml:"inscription"`
	Amount      decimal.Decimal `yaml:"amount"`
	NumTxs      int             `yaml:"numTxs"`
}

func loadConfig(filename string) (Config, error) {
	var config Config
	configFile, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func main() {
	// Load configuration from file
	config, err := loadConfig("config.yaml")
	if err != nil {
		fmt.Println("Load config file fail")
		fmt.Println(err)
	}

	// Validate required parameters
	if config.PrivateKey == "" || config.TransferTo == "" || config.Inscription == "" {
		fmt.Println("Please provide privateKey, transferTo, and inscription")
	}

	// 连接到以太坊节点
	client, err := ethclient.Dial(config.RPCURL) // 替换 YOUR-PROJECT-ID 为你的 Infura 项目 ID
	if err != nil {
		fmt.Printf("Connect %s fail\n", config.RPCURL)
		fmt.Println(err)
	}
	defer client.Close()

	privateKeyBytes, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		fmt.Println("Conver private key error")
		fmt.Println(err)
	}

	transferTo := common.HexToAddress(config.TransferTo)
	inscription := config.Inscription

	nonce, err := client.NonceAt(context.Background(), crypto.PubkeyToAddress(privateKeyBytes.PublicKey), nil)
	if err != nil {
		fmt.Println("Get nonce error")
		fmt.Println(err)
	}

	for i := 0; i < config.NumTxs; i++ {

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			fmt.Println("Get gasPrice error")
			fmt.Println(err)
		}

		msg := ethereum.CallMsg{
			To:   &transferTo,         // 目标地址，如果是合约调用，请替换为合约地址
			Data: []byte(inscription), // 交易数据，如果是合约调用，请替换为合约方法和参数
		}

		gasLimit, err := client.EstimateGas(context.Background(), msg)
		if err != nil {
			fmt.Println("Get gasLimit error")
			fmt.Println(err)
		}
		amountInDecimal := config.Amount.Mul(decimal.NewFromInt(1000000000000000000))
		amountInInt64 := big.NewInt(amountInDecimal.IntPart())
		tx := types.NewTransaction(nonce, transferTo, amountInInt64, gasLimit, gasPrice, []byte(inscription))

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			fmt.Println("Get chain id faild")
			fmt.Println(err)
		}

		signer := types.NewEIP155Signer(chainID)
		signedTx, err := types.SignTx(tx, signer, privateKeyBytes)
		if err != nil {
			fmt.Println("Sign TX error")
			fmt.Println(err)
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			fmt.Println("Send Transaction error")
			fmt.Println(err)
		}

		fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
		nonce++
		time.Sleep(5 * time.Second) // Add a delay between transactions
	}
}
