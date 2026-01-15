package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/ninja0404/pump-go-sdk/pkg/autofill"
	sdkconfig "github.com/ninja0404/pump-go-sdk/pkg/config"
	sdkrpc "github.com/ninja0404/pump-go-sdk/pkg/rpc"
	"github.com/ninja0404/pump-go-sdk/pkg/txbuilder"
	"github.com/ninja0404/pump-go-sdk/pkg/wallet"
)

// 示例：按 SOL 数量买入（buy_exact_sol_in），自动推导账户、自动补 ATA。
func main() {
	ctx := context.Background()

	const (
		mintStr      = "9PHN8hqogwssrHvC3K9UdWxz6o5H9FJaQJpjYHc9pump"
		spendableSol = uint64(4_000_000) // 允许花费的 SOL（lamports）
		minTokensOut = uint64(1)         // 至少获得的 token 数量（最小单位）
		rpcURL       = rpc.MainNetBeta_RPC
	)

	// 从环境变量读取私钥
	privateKeyB58 := os.Getenv("PUMP_PRIVATE_KEY")
	if privateKeyB58 == "" {
		log.Fatal("PUMP_PRIVATE_KEY environment variable is required")
	}

	mint := solana.MustPublicKeyFromBase58(mintStr)

	// 配置与客户端
	cfg := sdkconfig.DefaultRPCConfig()
	cfg.RPCURL = rpcURL
	cfg.Timeout = 20 * time.Second
	client := sdkrpc.NewClient(cfg)
	builder := txbuilder.NewBuilder(client, rpc.CommitmentProcessed)

	// 签名者
	signer, err := wallet.NewLocalFromBase58(privateKeyB58)
	if err != nil {
		log.Fatalf("load signer from base58: %v", err)
	}
	user := signer.PublicKey()

	dev := solana.MustPublicKeyFromBase58("BRFWRdf7ccq4pGnbyemruJUn7fkTL2kvekmJeoqspqX6")
	program := solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
	accounts, args, instrs, err := autofill.PumpEasyBuyExactSolIn(ctx, user, mint, dev, program, spendableSol, minTokensOut)
	if err != nil {
		log.Fatalf("autofill/build ix: %v", err)
	}

	tx, err := builder.BuildTransaction(ctx, signer.PublicKey(), instrs...)
	if err != nil {
		log.Fatalf("build tx: %v", err)
	}
	if err := txbuilder.SignTransaction(ctx, tx, signer); err != nil {
		log.Fatalf("sign: %v", err)
	}
	logs, err := builder.SendSimulate(ctx, tx)
	if err != nil {
		log.Fatalf("send: %v", err)
	}
	for _, v := range logs {
		fmt.Println(v)
	}
	_ = accounts
	_ = args
}
