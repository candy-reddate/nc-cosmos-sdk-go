package main

import (
	"fmt"
	"time"

	"github.com/irisnet/irismod-sdk-go/mt"
	"github.com/irisnet/irismod-sdk-go/nft"

	"github.com/irisnet/irismod-sdk-go/record"

	spartan "github.com/bianjieai/spartan-sdk-go/pkg/app/sdk"
	"github.com/irisnet/core-sdk-go/types"
	"github.com/irisnet/core-sdk-go/types/store"
	tendermintTypes "github.com/tendermint/tendermint/abci/types"
)

// Client initial info
var (
	rpcAddress  = "http://localhost:26657"
	grpcAddress = "localhost:9090"
	chainID     = "starmint"

	algo = "eth_secp256k1"

	// private key info
	name     = "mykey"
	password = "12345678"
	mnemonic = "hover acquire wave build cause wage mobile unlock thought never hint cup cricket again valve weekend voice session almost sadness erase february execute ripple"
)

func main() {
	// initial SDK configuration
	fee, _ := types.ParseDecCoins("200000ugas") // The default fee configuration
	options := []types.Option{
		types.AlgoOption(algo),
		types.KeyDAOOption(store.NewMemory(nil)),
		types.FeeOption(fee),
		types.TimeoutOption(10),
		types.CachedOption(true),
	}
	cfg, err := types.NewClientConfig(rpcAddress, grpcAddress, chainID, options...)
	if err != nil {
		panic(err)
	}

	// Create client
	client := spartan.NewClient(cfg, nil)

	// Private key import
	address, err := client.Key.Recover(name, password, mnemonic)
	if err != nil {
		fmt.Println(fmt.Errorf("Private key import failed: %s", err.Error()))
		return
	}
	fmt.Println("address:", address)

	// Initialize Tx arguments
	baseTx := types.BaseTx{
		From:     name,       // the same as the name of private key
		Password: password,   // the same as the password of private key
		Gas:      400000,     // the Gas limit of a single Tx
		Memo:     "",         // Tx Memo
		Mode:     types.Sync, // Tx broadcast mode
	}
	// Initialize Tx hash queue
	var hashArray []string

	// use Client to choose specific module to query the status info on chain
	acc, err := client.Bank.QueryAccount(address)
	if err != nil {
		fmt.Println(fmt.Errorf("the account query fail: %s", err.Error()))
	} else {
		fmt.Println("the account query success：", acc)
	}

	// use Client to choose a specific module to mint, sign and send Tx
	nftResult, err := client.NFT.IssueDenom(nft.IssueDenomRequest{ID: "testdenom", Name: "TestDenom", Schema: "{}"}, baseTx)
	if err != nil {
		fmt.Println(fmt.Errorf("NFT denom issue fail: %s", err.Error()))
	} else {
		fmt.Println("NFT denom issue success TxHash：", nftResult.Hash)
		hashArray = append(hashArray, nftResult.Hash)
	}

	// mint NFT
	mintNFT, err := client.NFT.MintNFT(nft.MintNFTRequest{Denom: "testdenom", ID: "testnft1", Name: "aaa", URI: "www.test.com", Data: "test", Recipient: address}, baseTx)
	if err != nil {
		e := err.(types.Error)
		if e.Codespace() == nft.ErrInvalidTokenID.Codespace() {
			fmt.Println("Err code: ", e.Code())
		}
		fmt.Println(fmt.Errorf("NFT mint fail: %s", err))
	} else {
		fmt.Println("NFT mint success TxHash：", mintNFT.Hash)
		hashArray = append(hashArray, mintNFT.Hash)
	}

	// use Client to choose a specific module to mint, sign and send Tx
	mtResult, err := client.MT.IssueDenom(mt.IssueDenomRequest{Name: "TestDenom", Data: []byte("TestData")}, baseTx)
	if err != nil {
		fmt.Println(fmt.Errorf("MT denom issue failed: %s", err.Error()))
	} else {
		fmt.Println("MT denom issue success TxHash：", mtResult.Hash)
		hashArray = append(hashArray, mtResult.Hash)
	}

	// use Client to choose a specific module to mint, sign and send Tx
	req := record.CreateRecordRequest{
		Contents: []record.Content{
			{
				Digest:     "digest", // meta data record
				DigestAlgo: "sha256", // meta data record algorithm
				URI:        "www.google.com",
				Meta:       "tx", // meta data
			},
		},
	}
	recordResult, err := client.Record.CreateRecord(req, baseTx)
	if err != nil {
		fmt.Println(fmt.Errorf("record creation failed: %s", err.Error()))
	} else {
		fmt.Println("record sending success：", recordResult.Hash)
		hashArray = append(hashArray, recordResult.Hash)
	}

	// wait 10 sec then query the Txs
	time.Sleep(time.Second * 10)
	for _, hash := range hashArray {
		tx, err := client.QueryTx(hash)
		if err != nil {
			fmt.Println("query Txs failed：", err)
			continue
		}
		if tx.Result.Code == tendermintTypes.CodeTypeOK {
			fmt.Println("Tx success, Tx hash:", hash)
		} else {
			fmt.Printf("Tx failed, Tx hash:%s， error info:%s. \n", hash, tx.Result.Log)
		}
	}

	// use Client to subscribe event notification
	subs, err := client.SubscribeNewBlock(types.NewEventQueryBuilder(), func(block types.EventDataNewBlock) {
		fmt.Println(block)
	})
	if err != nil {
		fmt.Println(fmt.Errorf("block subscribe failed: %s", err.Error()))
	} else {
		fmt.Println("block subscribe success：", subs.ID)
	}
	time.Sleep(time.Second * 20)
}
