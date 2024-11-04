package abci

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/cometbft/cometbft/abci/tutorials/abci-v2-forum-app/model"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptoencoding "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/version"
)

const (
	ApplicationVersion = 1
	CurseWordsLimitVE  = 10
)

type ForumApp struct {
	abci.BaseApplication
	valAddrToPubKeyMap map[string]crypto.PublicKey
	CurseWords         string
	state              AppState
	onGoingBlock       *badger.Txn
	logger             log.Logger
}

func NewForumApp(dbDir string, appConfigPath string, logger log.Logger) (*ForumApp, error) {
	db, err := model.NewDB(dbDir)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}
	cfg, err := LoadConfig(appConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %w", err)
	}

	cfg.CurseWords = DeduplicateCurseWords(cfg.CurseWords)

	state, err := loadState(db)
	if err != nil {
		return nil, err
	}

	// Reading the validators from the DB because CometBFT expects the application to have them in memory
	valMap := make(map[string]crypto.PublicKey)
	validators, err := state.DB.GetValidators()
	if err != nil {
		return nil, fmt.Errorf("can't load validators: %w", err)
	}
	for _, v := range validators {
		pubKey, err := cryptoencoding.PubKeyFromTypeAndBytes(v.PubKeyType, v.PubKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("can't decode public key: %w", err)
		}

		valMap[string(pubKey.Address())] = pubKey
	}

	return &ForumApp{
		state:              state,
		valAddrToPubKeyMap: valMap,
		CurseWords:         cfg.CurseWords,
		logger:             logger,
	}, nil
}

// Info return application information.
func (app *ForumApp) Info(_ context.Context, _ *abci.InfoRequest) (*abci.InfoResponse, error) {
	return &abci.InfoResponse{
		Version:         version.ABCIVersion,
		AppVersion:      ApplicationVersion,
		LastBlockHeight: app.state.Height,

		LastBlockAppHash: app.state.Hash(),
	}, nil
}

// Query the application state for specific information.
func (app *ForumApp) Query(_ context.Context, query *abci.QueryRequest) (*abci.QueryResponse, error) {
	app.logger.Info("Executing Application Query")

	resp := abci.QueryResponse{Key: query.Data}

	// Parse sender from query data
	sender := string(query.Data)

	if sender == "history" {
		messages, err := model.FetchHistory(app.state.DB)
		if err != nil {
			return nil, err
		}
		resp.Log = messages
		resp.Value = []byte(messages)

		return &resp, nil
	}
	// Retrieve all message sent by the sender
	messages, err := model.GetMessagesBySender(app.state.DB, sender)
	if err != nil {
		return nil, err
	}

	// Convert the messages to JSON and return as query result
	resultBytes, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	resp.Log = string(resultBytes)
	resp.Value = resultBytes

	return &resp, nil
}

// CheckTx handles validation of inbound transactions. If a transaction is not a valid message, or if a user
// does not exist in the database or if a user is banned it returns an error.
func (app *ForumApp) CheckTx(_ context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
	app.logger.Info("Executing Application CheckTx")

	// Parse the tx message
	msg, err := model.ParseMessage(req.Tx)
	if err != nil {
		app.logger.Info("CheckTx: failed to parse transaction message", "message", msg, "error", err)
		return &abci.CheckTxResponse{Code: CodeTypeInvalidTxFormat, Log: "Invalid transaction", Info: err.Error()}, nil
	}

	// Check for invalid sender
	if len(msg.Sender) == 0 {
		app.logger.Info("CheckTx: failed to parse transaction message", "message", msg, "error", "Sender is missing")
		return &abci.CheckTxResponse{Code: CodeTypeInvalidTxFormat, Log: "Invalid transaction", Info: "Sender is missing"}, nil
	}

	app.logger.Debug("searching for sender", "sender", msg.Sender)
	u, err := app.state.DB.FindUserByName(msg.Sender)

	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			app.logger.Error("CheckTx: Error in check tx", "tx", string(req.Tx), "error", err)
			return &abci.CheckTxResponse{Code: CodeTypeEncodingError, Log: "Invalid transaction", Info: err.Error()}, nil
		}
		app.logger.Info("CheckTx: Sender not found", "sender", msg.Sender)
	} else if u != nil && u.Banned {
		return &abci.CheckTxResponse{Code: CodeTypeBanned, Log: "Invalid transaction", Info: "User is banned"}, nil
	}
	app.logger.Info("CheckTx: success checking tx", "message", msg.Message, "sender", msg.Sender)
	return &abci.CheckTxResponse{Code: CodeTypeOK, Log: "Valid transaction", Info: "Transaction validation succeeded"}, nil
}

// Consensus Connection

// InitChain initializes the blockchain with information sent from CometBFT such as validators or consensus parameters.
func (app *ForumApp) InitChain(_ context.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	app.logger.Info("Executing Application InitChain")

	for _, v := range req.Validators {
		err := app.updateValidator(v)
		if err != nil {
			return nil, err
		}
	}
	appHash := app.state.Hash()

	// This parameter can also be set in the genesis file
	req.ConsensusParams.Feature.VoteExtensionsEnableHeight.Value = 1
	return &abci.InitChainResponse{ConsensusParams: req.ConsensusParams, AppHash: appHash}, nil
}

// PrepareProposal is used to prepare a proposal for the next block in the blockchain. The application can re-order, remove
// or add transactions.
func (app *ForumApp) PrepareProposal(_ context.Context, req *abci.PrepareProposalRequest) (*abci.PrepareProposalResponse, error) {
	app.logger.Info("Executing Application PrepareProposal")

	// Get the curse words from for all vote extensions received at the end of last height.
	voteExtensionCurseWords := app.getWordsFromVe(req.LocalLastCommit.Votes)

	curseWords := strings.Split(voteExtensionCurseWords, "|")
	if hasDuplicateWords(curseWords) {
		return nil, errors.New("duplicate words found")
	}

	// Prepare req puts the BanTx first, then adds the other transactions
	// ProcessProposal should verify this
	proposedTxs := make([][]byte, 0)
	finalProposal := make([][]byte, 0)
	bannedUsersString := make(map[string]struct{})
	for _, tx := range req.Txs {
		msg, err := model.ParseMessage(tx)
		if err != nil {
			// this should never happen since the tx should have been validated by CheckTx
			return nil, fmt.Errorf("failed to marshal tx in PrepareProposal: %w", err)
		}
		// Adding the curse words from vote extensions too
		if !hasCurseWord(msg.Message, voteExtensionCurseWords) {
			proposedTxs = append(proposedTxs, tx)
			continue
		}
		// If the message contains curse words then ban the user by
		// creating a "ban transaction" and adding it to the final proposal
		banTx := model.BanTx{UserName: msg.Sender}
		bannedUsersString[msg.Sender] = struct{}{}
		resultBytes, err := json.Marshal(banTx)
		if err != nil {
			// this should never happen since the ban tx should have been validated by CheckTx
			return nil, fmt.Errorf("failed to marshal ban tx in PrepareProposal: %w", err)
		}
		finalProposal = append(finalProposal, resultBytes)
	}

	// Need to loop again through the proposed Txs to make sure there is none left by a user that was banned
	// after the tx was accepted
	for _, tx := range proposedTxs {
		// there should be no error here as these are just transactions we have checked and added
		msg, err := model.ParseMessage(tx)
		if err != nil {
			// this should never happen since the tx should have been validated by CheckTx
			return nil, fmt.Errorf("failed to marshal tx in PrepareProposal: %w", err)
		}
		// If the user is banned then include this transaction in the final proposal
		if _, ok := bannedUsersString[msg.Sender]; !ok {
			finalProposal = append(finalProposal, tx)
		}
	}
	return &abci.PrepareProposalResponse{Txs: finalProposal}, nil
}

// ProcessProposal validates the proposed block and the transactions and return a status if it was accepted or rejected.
func (app *ForumApp) ProcessProposal(_ context.Context, req *abci.ProcessProposalRequest) (*abci.ProcessProposalResponse, error) {
	app.logger.Info("Executing Application ProcessProposal")

	bannedUsers := make(map[string]struct{}, 0)

	finishedBanTxIdx := len(req.Txs)
	for i, tx := range req.Txs {
		if !isBanTx(tx) {
			finishedBanTxIdx = i
			break
		}
		var parsedBan model.BanTx
		err := json.Unmarshal(tx, &parsedBan)
		if err != nil {
			return &abci.ProcessProposalResponse{Status: abci.PROCESS_PROPOSAL_STATUS_REJECT}, err
		}
		bannedUsers[parsedBan.UserName] = struct{}{}
	}

	for _, tx := range req.Txs[finishedBanTxIdx:] {
		// From this point on, there should be no BanTxs anymore
		// If there is one, ParseMessage will return an error as the
		// format of the two transactions is different.
		msg, err := model.ParseMessage(tx)
		if err != nil {
			return &abci.ProcessProposalResponse{Status: abci.PROCESS_PROPOSAL_STATUS_REJECT}, err
		}
		if _, ok := bannedUsers[msg.Sender]; ok {
			// sending us a tx from a banned user
			return &abci.ProcessProposalResponse{Status: abci.PROCESS_PROPOSAL_STATUS_REJECT}, nil
		}
	}
	return &abci.ProcessProposalResponse{Status: abci.PROCESS_PROPOSAL_STATUS_ACCEPT}, nil
}

// FinalizeBlock Deliver the decided block to the Application.
func (app *ForumApp) FinalizeBlock(_ context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
	app.logger.Info("Executing Application FinalizeBlock")

	// Iterate over Tx in current block
	app.onGoingBlock = app.state.DB.GetDB().NewTransaction(true)
	respTxs := make([]*abci.ExecTxResult, len(req.Txs))
	finishedBanTxIdx := len(req.Txs)
	for i, tx := range req.Txs {
		var err error

		if !isBanTx(tx) {
			finishedBanTxIdx = i
			break
		}
		banTx := new(model.BanTx)
		err = json.Unmarshal(tx, &banTx)
		if err != nil {
			// since we did this in ProcessProposal this should never happen here
			return nil, err
		}
		err = UpdateOrSetUser(app.state.DB, banTx.UserName, true, app.onGoingBlock)
		if err != nil {
			return nil, err
		}
		respTxs[i] = &abci.ExecTxResult{Code: CodeTypeOK}
	}

	for idx, tx := range req.Txs[finishedBanTxIdx:] {
		// From this point on, there should be no BanTxs anymore
		// If there is one, ParseMessage will return an error as the
		// format of the two transactions is different.
		msg, err := model.ParseMessage(tx)
		i := idx + finishedBanTxIdx
		if err != nil {
			// since we did this in ProcessProposal this should never happen here
			return nil, err
		}

		// Check if this sender already existed; if not, add the user too
		err = UpdateOrSetUser(app.state.DB, msg.Sender, false, app.onGoingBlock)
		if err != nil {
			return nil, err
		}
		// Add the message for this sender
		message, err := model.AppendToExistingMessages(app.state.DB, *msg)
		if err != nil {
			return nil, err
		}
		err = app.onGoingBlock.Set([]byte(msg.Sender+"msg"), []byte(message))
		if err != nil {
			return nil, err
		}
		chatHistory, err := model.AppendToChat(app.state.DB, *msg)
		if err != nil {
			return nil, err
		}
		// Append messages to chat history
		err = app.onGoingBlock.Set([]byte("history"), []byte(chatHistory))
		if err != nil {
			return nil, err
		}
		// This adds the user to the DB, but the data is not committed nor persisted until Commit is called
		respTxs[i] = &abci.ExecTxResult{Code: abci.CodeTypeOK}
		app.state.Size++
	}
	app.state.Height = req.Height

	response := &abci.FinalizeBlockResponse{TxResults: respTxs, AppHash: app.state.Hash(), NextBlockDelay: 1 * time.Second}
	return response, nil
}

// Commit the application state.
func (app *ForumApp) Commit(_ context.Context, _ *abci.CommitRequest) (*abci.CommitResponse, error) {
	app.logger.Info("Executing Application Commit")

	if err := app.onGoingBlock.Commit(); err != nil {
		return nil, err
	}
	err := saveState(&app.state)
	if err != nil {
		return nil, err
	}
	return &abci.CommitResponse{}, nil
}

// ExtendVote returns curse words as vote extensions.
func (app *ForumApp) ExtendVote(_ context.Context, _ *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	app.logger.Info("Executing Application ExtendVote")

	return &abci.ExtendVoteResponse{VoteExtension: []byte(app.CurseWords)}, nil
}

// VerifyVoteExtension verifies the vote extensions and ensure they include the curse words
// It will not be called for extensions generated by this validator.
func (app *ForumApp) VerifyVoteExtension(_ context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	app.logger.Info("Executing Application VerifyVoteExtension")

	if _, ok := app.valAddrToPubKeyMap[string(req.ValidatorAddress)]; !ok {
		// we do not have a validator with this address mapped; this should never happen
		return nil, errors.New("unknown validator")
	}

	curseWords := strings.Split(string(req.VoteExtension), "|")
	if hasDuplicateWords(curseWords) {
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	// ensure vote extension curse words limit has not been exceeded
	if len(curseWords) > CurseWordsLimitVE {
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}
	return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
}

// getWordsFromVE gets the curse words from the vote extensions as one string, the words are concatenated using a '|'
// this method also ensures there are no duplicate curse words in the final set returned.
func (app *ForumApp) getWordsFromVe(voteExtensions []abci.ExtendedVoteInfo) string {
	curseWordMap := make(map[string]int)
	for _, vote := range voteExtensions {
		// This code gets the curse words and makes sure that we do not add them more than once
		// Thus ensuring each validator only adds one word once
		curseWords := strings.Split(string(vote.GetVoteExtension()), "|")

		for _, word := range curseWords {
			if count, ok := curseWordMap[word]; !ok {
				curseWordMap[word] = 1
			} else {
				curseWordMap[word] = count + 1
			}
		}
	}
	app.logger.Info("Processed vote extensions", "curse_words", curseWordMap)
	majority := len(app.valAddrToPubKeyMap) / 3 // We define the majority to be at least 1/3 of the validators;

	voteExtensionCurseWords := ""
	for word, count := range curseWordMap {
		if count > majority {
			if voteExtensionCurseWords == "" {
				voteExtensionCurseWords = word
			} else {
				voteExtensionCurseWords += "|" + word
			}
		}
	}
	return voteExtensionCurseWords
}

// hasDuplicateWords detects if there are duplicate words in the slice.
func hasDuplicateWords(words []string) bool {
	wordMap := make(map[string]struct{})

	for _, word := range words {
		wordMap[word] = struct{}{}
	}

	return len(words) != len(wordMap)
}
