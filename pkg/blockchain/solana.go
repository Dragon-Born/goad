package blockchain

import (
	"context"
	"errors"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/assotokenprog"
	"github.com/blocto/solana-go-sdk/program/cmptbdgprog"
	"github.com/blocto/solana-go-sdk/program/memo"
	"github.com/blocto/solana-go-sdk/program/metaplex/tokenmeta"
	"github.com/blocto/solana-go-sdk/program/sysprog"
	"github.com/blocto/solana-go-sdk/program/tokenprog"
	"github.com/blocto/solana-go-sdk/types"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
)

const mainnet = "https://api.mainnet-beta.solana.com"

type SolClient struct {
	privateKey   solana.PrivateKey
	nonceAccount common.PublicKey
	tokenBalance map[string]float64
	SOLBalance   float64
	client       *client.Client
}

func (a *SolClient) TokenBalance(symbol string) (balance float64) {
	balance, ok := a.tokenBalance[symbol]
	if !ok {
		balance = 0
	}
	return
}

func (a *SolClient) SetTokenBalance(symbol string, balance float64) bool {
	if a.tokenBalance[symbol] == balance {
		return false
	}
	a.tokenBalance[symbol] = balance
	return true
}

func (s *SolClient) Address() common.PublicKey {
	return common.PublicKeyFromString(s.privateKey.PublicKey().String())
}

func NewSolanaWallet(priKey, nonceAccount string) (*SolClient, error) {
	sender, err := solana.PrivateKeyFromBase58(priKey)
	if err != nil {
		return nil, err
	}
	return &SolClient{privateKey: sender, nonceAccount: common.PublicKeyFromString(nonceAccount), tokenBalance: make(map[string]float64)}, nil
}

func NewSolanaClient(url string) *SolClient {
	return &SolClient{
		client: client.NewClient(url),
	}
}

func (w *SolClient) GetCurrentBlock() (block *client.Block, err error) {
	c := client.NewClient(mainnet)
	ctx := context.Background()
	latestSlot, err := c.GetSlot(ctx)
	if err != nil {
		fmt.Println("Failed to fetch latest block: ", err)
		return nil, err
	}
	block, err = c.GetBlock(ctx, latestSlot)
	if err != nil {
		fmt.Println("Failed to fetch the latest confirmed block: ", err)
		return nil, err
	}
	return block, nil
}

func GetTokenInfoByMintAddress(mintAddress string) (*tokenmeta.Metadata, error) {
	mint := common.PublicKeyFromString(mintAddress)
	metadataAccount, err := tokenmeta.GetTokenMetaPubkey(mint)
	if err != nil {
		return nil, err
	}

	// new a client
	c := client.NewClient(mainnet)

	// get data which stored in metadataAccount
	accountInfo, err := c.GetAccountInfo(context.Background(), metadataAccount.ToBase58())
	if err != nil {
		return nil, err
	}

	// parse it
	metadata, err := tokenmeta.MetadataDeserialize(accountInfo.Data)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func GetBlockTransactions(block *client.Block) (dexTransactions []SolDexTransaction, err error) {
	raydiumProgramID := "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	//fmt.Printf("Number of Transactions: %d \n\n", len(block.Transactions))
	i := 0
	for _, tx := range block.Transactions {
		if tx.Meta != nil && tx.Meta.Err != nil {
			continue
		}
		isRaydiumTx := false
		for _, instruction := range tx.Transaction.Message.Instructions {
			// Retrieve the program ID using the index
			programID := tx.Transaction.Message.Accounts[instruction.ProgramIDIndex]

			// Check if the program ID matches Raydium's program ID
			if programID.String() == raydiumProgramID {
				isRaydiumTx = true
				break
			}
		}
		if !isRaydiumTx {
			continue
		}
		//fmt.Printf("Transaction #%d:\n", i+1)
		i++
		//fmt.Printf("  Signature: %s\n", tx.Transaction.Signatures[0])
		if len(tx.Transaction.Signatures) > 0 {
			// Print the first signature (considered the sender's signature)
			//if len(tx.Transaction.Signatures) > 0 {
			//	firstSignature := tx.Transaction.Signatures[0]
			//	fmt.Printf("  TX: %s\n", base58.Encode(firstSignature[:]))
			//} else {
			//	fmt.Println("  No Signatures")
			//}

			// Extract the fee payer (sender)
			if len(tx.Transaction.Message.Accounts) > 0 {
				senderAccount := tx.Transaction.Message.Accounts[0]
				dexTransactions = append(dexTransactions, SolDexTransaction{
					From:     senderAccount,
					Method:   "DEX",
					Amount:   nil,
					GasPrice: nil,
				})
				//fmt.Printf("  Wallet: %s\n", senderAccount.String())
			} else {
				//fmt.Println("  No Accounts")
			}
			//if tx.Meta != nil {
			//	preTokenBalances := tx.Meta.PreTokenBalances
			//	postTokenBalances := tx.Meta.PostTokenBalances
			//
			//	// Show all affected token accounts and their changes
			//	for j := range preTokenBalances {
			//		pre := preTokenBalances[j]
			//		post := postTokenBalances[j]
			//
			//		fmt.Printf("  Token Account: %s\n", post.Owner)
			//		fmt.Printf("    Mint: %s\n", post.Mint)
			//		parseIntPost, err := strconv.ParseInt(post.UITokenAmount.Amount, 10, 64)
			//		if err != nil {
			//			continue
			//		}
			//		parseIntPre, err := strconv.ParseInt(pre.UITokenAmount.Amount, 10, 64)
			//		if err != nil {
			//			continue
			//		}
			//		fmt.Printf("    Pre-Amount: %f\n", float64(parseIntPre)/math.Pow10(int(pre.UITokenAmount.Decimals)))
			//		fmt.Printf("    Post-Amount: %f\n\n", float64(parseIntPost)/math.Pow10(int(post.UITokenAmount.Decimals)))
			//	}
			//}
		}

		// Include other transaction details as needed

	}
	return dexTransactions, nil
}

func IsSolanaWallet(address string) bool {
	_, err := solana.PublicKeyFromBase58(address)
	return err == nil
}

func GetSolBalance(publicKey common.PublicKey) (float64, error) {
	_pubKey, err := solana.PublicKeyFromBase58(publicKey.ToBase58())
	if err != nil {
		return 0, err
	}
	c := client.NewClient(mainnet)
	//time.Sleep(5 * time.Second)

	balance, err := c.GetBalance(
		context.Background(),
		_pubKey.String())
	if err != nil {
		return 0, err
	}
	return float64(balance) / float64(solana.LAMPORTS_PER_SOL), nil
}

func (w *SolClient) CreateAndInitializeNonceAccount() (common.PublicKey, error) {
	feePayer, err := types.AccountFromBase58(w.privateKey.String())
	if err != nil {
		return common.PublicKey{}, err
	}
	c := client.NewClient(mainnet)
	//time.Sleep(5 * time.Second)
	nonceAccountRentFreeBalance, err := c.GetMinimumBalanceForRentExemption(
		context.Background(),
		sysprog.NonceAccountSize,
	)

	if err != nil {
		return common.PublicKey{}, err
	}
	nonceAccount := types.NewAccount()
	fmt.Println("nonce account:", nonceAccount.PublicKey)
	instructions := []types.Instruction{
		sysprog.CreateAccount(sysprog.CreateAccountParam{
			From:     feePayer.PublicKey,
			New:      nonceAccount.PublicKey,
			Owner:    common.SystemProgramID,
			Lamports: nonceAccountRentFreeBalance,
			Space:    sysprog.NonceAccountSize,
		}),
		sysprog.InitializeNonceAccount(sysprog.InitializeNonceAccountParam{
			Nonce: nonceAccount.PublicKey,
			Auth:  feePayer.PublicKey,
		}),
	}
	//time.Sleep(5 * time.Second)
	blockhash, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		return common.PublicKey{}, err
	}
	// Create the transaction message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        feePayer.PublicKey,
		RecentBlockhash: blockhash.Blockhash,
		Instructions:    instructions,
	})

	// Create the transaction
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, nonceAccount}, // include all necessary signers
	})
	_, err = tx.Serialize()
	if err != nil {
		return common.PublicKey{}, err
	}
	//time.Sleep(5 * time.Second)

	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		return common.PublicKey{}, err
	}
	fmt.Println("txhash: ", txhash)
	return nonceAccount.PublicKey, nil
	//signedTx, err := tx.Sign(func(pubKey common.PublicKey) *types.Account {
	//	if pubKey.Equals(feePayer.PublicKey) {
	//		return &feePayer
	//	} else if pubKey.Equals(nonceAccount.PublicKey) {
	//		return &nonceAccount
	//	}
	//	return nil
	//})
}

func (w *SolClient) GetTX(hash string) (*types.Transaction, error) {
	c := client.NewClient(mainnet)
	tx, err := c.GetTransaction(context.Background(), hash)
	if err != nil {
		return nil, err
	}
	if tx == nil || tx.BlockTime == nil {
		return nil, nil
	}
	return &tx.Transaction, nil
}

func (w *SolClient) SendSLPToken(tokenAddr, receiver string, amount float64) (string, error) {
	c := client.NewClient(mainnet)
	tokenAddrPub := common.PublicKeyFromString(tokenAddr)
	feePayer, err := types.AccountFromBase58(w.privateKey.String())
	if err != nil {
		return "", errors.Join(errors.New("failed to parse private key"), err)
	}
	alice := feePayer
	aliceTA, _, err := common.FindAssociatedTokenAddress(alice.PublicKey, tokenAddrPub)
	if err != nil {
		return "", errors.Join(errors.New("failed to find associated token address"), err)
	}
	receiverAddrPub := common.PublicKeyFromString(receiver)

	receiverTA, _, err := common.FindAssociatedTokenAddress(receiverAddrPub, tokenAddrPub)
	if err != nil {
		return "", errors.Join(errors.New("failed to find associated token address"), err)
	}
	//fmt.Println("receiver token address: ", receiverTA.String())
	accountInfo, err := c.GetAccountInfo(context.Background(), receiverTA.ToBase58())
	if err != nil {
		return "", errors.Join(errors.New("failed to get account info"), err)
	}

	var instructions []types.Instruction
	//nonceAccountInfo, err := c.GetAccountInfo(
	//	context.Background(),
	//	w.nonceAccount.ToBase58(),
	//)
	//if err != nil {
	//	return "", errors.Join(errors.New("failed to get nonce account info"), err)
	//}
	//
	//nonceAccount, err := sysprog.NonceAccountDeserialize(nonceAccountInfo.Data)
	//if err != nil {
	//	return "", errors.Join(errors.New("failed to deserialize nonce account"), err)
	//}
	//
	//nonceAccountPubKey := w.nonceAccount

	//instructions = append(instructions, sysprog.AdvanceNonceAccount(sysprog.AdvanceNonceAccountParam{
	//	Nonce: nonceAccountPubKey,
	//	Auth:  feePayer.PublicKey,
	//}))
	instructions = append(instructions,
		//	cmptbdgprog.SetComputeUnitLimit(
		//	cmptbdgprog.SetComputeUnitLimitParam{
		//		Units: 1_000_000,
		//	},
		//),
		cmptbdgprog.SetComputeUnitPrice(
			cmptbdgprog.SetComputeUnitPriceParam{
				MicroLamports: 5000,
			},
		),
	)
	if len(accountInfo.Data) == 0 {
		// Account does not exist or is uninitialized, include the create instruction
		instructions = append(instructions, assotokenprog.CreateAssociatedTokenAccount(assotokenprog.CreateAssociatedTokenAccountParam{
			Funder:                 feePayer.PublicKey,
			Owner:                  receiverAddrPub,
			Mint:                   tokenAddrPub,
			AssociatedTokenAccount: receiverTA,
		}))
	}
	instructions = append(instructions, tokenprog.TransferChecked(tokenprog.TransferCheckedParam{
		From:     aliceTA,
		To:       receiverTA,
		Mint:     tokenAddrPub,
		Auth:     alice.PublicKey,
		Signers:  []common.PublicKey{},
		Amount:   uint64(amount * float64(solana.LAMPORTS_PER_SOL)),
		Decimals: uint8(math.Log10(float64(solana.LAMPORTS_PER_SOL))),
	}))

	instructions = append(instructions,
		memo.BuildMemo(memo.BuildMemoParam{
			SignerPubkeys: []common.PublicKey{alice.PublicKey},
			Memo:          []byte("fuk da sistm, -> https://x.com/jinpengsol"),
		}),
	)
	latestHash, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		return "", errors.Join(errors.New("failed to get latest blockhash"), err)
	}
	message := types.NewMessage(types.NewMessageParam{
		FeePayer: feePayer.PublicKey,
		//RecentBlockhash: nonceAccount.Nonce.ToBase58(), // recent blockhash\
		RecentBlockhash: latestHash.Blockhash, // recent blockhash\
		Instructions:    instructions,
	})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, alice},
	})
	if err != nil {
		return "", errors.Join(errors.New("failed to create transaction"), err)
	}
	//time.Sleep(5 * time.Second)

	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", errors.Join(errors.New("failed to send transaction"), err)
	}
	return txhash, nil

}

func (w *SolClient) SendSOL(receiver string, amount float64) (string, error) {
	c := client.NewClient(mainnet)
	feePayer, err := types.AccountFromBase58(w.privateKey.String())
	if err != nil {
		return "", err
	}
	alice := feePayer
	// to fetch recent blockhash
	//time.Sleep(5 * time.Second)

	res, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		return "", err
	}

	// create a message
	message := types.NewMessage(types.NewMessageParam{
		FeePayer:        feePayer.PublicKey,
		RecentBlockhash: res.Blockhash, // recent blockhash
		Instructions: []types.Instruction{
			sysprog.Transfer(sysprog.TransferParam{
				From:   alice.PublicKey,                      // from
				To:     common.PublicKeyFromString(receiver), // to
				Amount: uint64(amount * float64(solana.LAMPORTS_PER_SOL)),
			}),
		},
	})

	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{feePayer, alice},
	})
	if err != nil {
		return "", err
	}

	// send tx
	//time.Sleep(5 * time.Second)

	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", err
	}
	return txhash, nil
}

func GetSolTokenBalance(publicKey common.PublicKey, address string) (float64, error) {
	_pubKey, err := solana.PublicKeyFromBase58(publicKey.ToBase58())
	if err != nil {
		return 0, err
	}
	mintAddress, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, err
	}
	c := rpc.New(mainnet)
	config := &rpc.GetTokenAccountsConfig{
		Mint: &mintAddress, // Filter accounts by the token mint address
	}
	opts := &rpc.GetTokenAccountsOpts{
		Commitment: rpc.CommitmentFinalized, // You can set the commitment level here
		Encoding:   solana.EncodingBase64,   // This is the default and recommended setting
	}
	//time.Sleep(5 * time.Second)
	// Fetch token accounts by owner:
	tokenAccounts, err := c.GetTokenAccountsByOwner(
		context.TODO(),
		_pubKey,
		config,
		opts,
	)
	if err != nil {
		return 0, err
	}
	if len(tokenAccounts.Value) > 0 {
		dec := bin.NewBinDecoder(tokenAccounts.Value[0].Account.Data.GetBinary())
		if dec == nil {
			return 0, nil
		}
		var tokenAccount token.Account
		if err = tokenAccount.UnmarshalWithDecoder(dec); err != nil {
			return 0, err
		}
		return float64(tokenAccount.Amount) / float64(solana.LAMPORTS_PER_SOL), nil
	}
	return 0, fmt.Errorf("no token account found")

}
