package abci

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
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
	// MaxMetricsCount is the maximum number of metrics allowed in a vote extension
	MaxMetricsCount = 10
)

// ForumApp is the main application implementing the ABCI interface for a forum application.
type ForumApp struct {
	abci.BaseApplication
	// valAddrToPubKeyMap maps validator addresses to their public keys
	valAddrToPubKeyMap map[string]crypto.PublicKey
	// CurseWords contains the list of words that are banned in the forum
	CurseWords string
	// NetworkMetrics stores the aggregated performance metrics from validators
	NetworkMetrics map[string]float64
	// MessageRateLimit defines how many messages a user can send per block
	// This limit is dynamically adjusted based on network performance
	MessageRateLimit int
	// state contains the application state
	state AppState
	// onGoingBlock is the transaction for the current block being processed
	onGoingBlock *badger.Txn
	// logger is used for application logging
	logger log.Logger
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
		NetworkMetrics:     make(map[string]float64),
		MessageRateLimit:   10,
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

	// Collect network metrics from vote extensions
	networkMetrics := app.getMetricsFromVoteExtensions(req.LocalLastCommit.Votes)

	// Adapt processing speed and limits based on network performance metrics
	app.adjustForumParameters(networkMetrics)

	// Prepare proposal with new parameters
	proposedTxs := make([][]byte, 0)
	finalProposal := make([][]byte, 0)
	bannedUsersString := make(map[string]struct{})
	userMessageCounts := make(map[string]int)

	for _, tx := range req.Txs {
		msg, err := model.ParseMessage(tx)
		if err != nil {
			// this should never happen since the tx should have been validated by CheckTx
			return nil, fmt.Errorf("failed to marshal tx in PrepareProposal: %w", err)
		}

		// Check for forbidden words
		if hasCurseWord(msg.Message, app.CurseWords) {
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
			continue
		}

		// Control message rate from users based on current limit
		userMessageCounts[msg.Sender]++
		if userMessageCounts[msg.Sender] <= app.MessageRateLimit {
			proposedTxs = append(proposedTxs, tx)
		}
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

// ExtendVote returns validator metrics as vote extensions
func (app *ForumApp) ExtendVote(_ context.Context, _ *abci.ExtendVoteRequest) (*abci.ExtendVoteResponse, error) {
	app.logger.Info("Executing Application ExtendVote")

	// Collect metrics to include in the vote extension
	metrics := map[string]float64{
		"cpu_usage":     getCPUUsage(),        // Example function that returns CPU usage
		"memory_usage":  getMemoryUsage(),     // Example function that returns memory usage
		"disk_io":       getDiskIO(),          // Example function that returns disk activity
		"message_count": getMessageCount(app), // Function that returns the number of messages in the block
		"timestamp":     float64(time.Now().Unix()),
	}

	// Serialize metrics to JSON
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return &abci.ExtendVoteResponse{VoteExtension: []byte("{}")}, nil
	}

	return &abci.ExtendVoteResponse{VoteExtension: metricsBytes}, nil
}

// VerifyVoteExtension verifies the vote extensions and ensures they contain valid metrics
func (app *ForumApp) VerifyVoteExtension(_ context.Context, req *abci.VerifyVoteExtensionRequest) (*abci.VerifyVoteExtensionResponse, error) {
	app.logger.Info("Executing Application VerifyVoteExtension")

	if _, ok := app.valAddrToPubKeyMap[string(req.ValidatorAddress)]; !ok {
		// we do not have a validator with this address mapped; this should never happen
		return nil, errors.New("unknown validator")
	}

	// Try to decode metrics from vote extension
	var metrics map[string]float64
	err := json.Unmarshal(req.VoteExtension, &metrics)
	if err != nil {
		app.logger.Info("Invalid metrics format in vote extension", "error", err)
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	// Check for required fields
	requiredMetrics := []string{"cpu_usage", "memory_usage", "timestamp"}
	for _, metricName := range requiredMetrics {
		if _, exists := metrics[metricName]; !exists {
			app.logger.Info("Missing required metric in vote extension", "metric", metricName)
			return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
		}
	}

	// Validate metric values for reasonableness
	if metrics["cpu_usage"] < 0 || metrics["cpu_usage"] > 100 {
		app.logger.Info("Invalid CPU usage value", "value", metrics["cpu_usage"])
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	if metrics["memory_usage"] < 0 {
		app.logger.Info("Invalid memory usage value", "value", metrics["memory_usage"])
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	// Check that timestamp is not from the future and not too old
	now := float64(time.Now().Unix())
	if metrics["timestamp"] > now+60 { // Allow up to 1 minute difference
		app.logger.Info("Timestamp is from the future", "timestamp", metrics["timestamp"], "now", now)
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	if now-metrics["timestamp"] > 300 { // Reject metrics older than 5 minutes
		app.logger.Info("Timestamp is too old", "timestamp", metrics["timestamp"], "now", now)
		return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_REJECT}, nil
	}

	return &abci.VerifyVoteExtensionResponse{Status: abci.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT}, nil
}

// getMetricsFromVoteExtensions collects metrics from vote extensions
func (app *ForumApp) getMetricsFromVoteExtensions(voteExtensions []abci.ExtendedVoteInfo) map[string]float64 {
	combinedMetrics := make(map[string]float64)
	validatorCount := 0

	for _, vote := range voteExtensions {
		if len(vote.GetVoteExtension()) == 0 {
			continue
		}

		var metrics map[string]float64
		err := json.Unmarshal(vote.GetVoteExtension(), &metrics)
		if err != nil {
			app.logger.Info("Failed to unmarshal metrics from vote extension", "error", err)
			continue
		}

		// Perform aggregation for each metric
		for name, value := range metrics {
			if name == "timestamp" {
				continue // Skip timestamp, it's only needed for validation
			}

			if existing, ok := combinedMetrics[name]; ok {
				combinedMetrics[name] = existing + value
			} else {
				combinedMetrics[name] = value
			}
		}

		validatorCount++
	}

	// Calculate average values for all metrics
	if validatorCount > 0 {
		for name, total := range combinedMetrics {
			combinedMetrics[name] = total / float64(validatorCount)
		}
	}

	app.logger.Info("Processed vote extensions metrics", "metrics", combinedMetrics, "validator_count", validatorCount)
	return combinedMetrics
}

// adjustForumParameters adapts forum parameters based on network metrics
func (app *ForumApp) adjustForumParameters(metrics map[string]float64) {
	// Save metrics for later use
	app.NetworkMetrics = metrics

	// Adjust message rate limits based on network load
	cpuUsage, hasCPU := metrics["cpu_usage"]
	memoryUsage, hasMemory := metrics["memory_usage"]

	// If metrics are available, adjust parameters
	if hasCPU && hasMemory {
		// High CPU or memory usage -> decrease message limit
		if cpuUsage > 80 || memoryUsage > 80 {
			app.MessageRateLimit = 5
		} else if cpuUsage > 60 || memoryUsage > 60 {
			app.MessageRateLimit = 8
		} else {
			// Low load -> increase message limit
			app.MessageRateLimit = 15
		}
	}

	app.logger.Info("Adjusted forum parameters based on metrics",
		"message_rate_limit", app.MessageRateLimit,
		"cpu_usage", cpuUsage,
		"memory_usage", memoryUsage)
}

// getCPUUsage returns the current CPU usage as a percentage.
// In a real implementation, this would measure actual system metrics.
func getCPUUsage() float64 {
	// In a real application, there would be code here to measure CPU usage
	// For example purposes, we return a fixed value
	return 50.0 // CPU usage percentage
}

// getMemoryUsage returns the current memory usage as a percentage.
// In a real implementation, this would measure actual system metrics.
func getMemoryUsage() float64 {
	// In a real application, there would be code here to measure memory usage
	// For example purposes, we return a fixed value
	return 40.0 // Memory usage percentage
}

// getDiskIO returns the current disk I/O activity.
// In a real implementation, this would measure actual system metrics.
func getDiskIO() float64 {
	// In a real application, there would be code here to measure disk activity
	// For example purposes, we return a fixed value
	return 30.0 // Operations per second or other IO metric
}

// getMessageCount returns the number of messages processed in the current block.
// In a real implementation, this would count actual messages.
func getMessageCount(app *ForumApp) float64 {
	// In a real application, there would be code here to count messages
	// For example purposes, we return a fixed value
	return 25.0 // Number of messages
}

// hasDuplicateWords detects if there are duplicate words in the slice.
func hasDuplicateWords(words []string) bool {
	wordMap := make(map[string]struct{})

	for _, word := range words {
		wordMap[word] = struct{}{}
	}

	return len(words) != len(wordMap)
}
