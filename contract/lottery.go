package main

import (
	"okinoko_lottery/sdk"
	"strconv"
)

//export create_lottery
func create_lottery(payload *string) *string {
	payloadStr := unwrapPayload(payload, "create_lottery payload missing")
	args := parseCreateLottery(payloadStr)

	// No transfer intent needed for creation
	sender := getSenderAddress()
	now := nowUnix()

	// Create new lottery
	lottery := &Lottery{
		ID:           getNextLotteryID(),
		Creator:      sender,
		Name:         args.Name,
		CreatedAt:    now,
		DeadlineDays: args.DeadlineDays,
		DeadlineUnix: now + int64(args.DeadlineDays*24*60*60),
		BurnPercent:  args.BurnPercent,
		TicketPrice:  args.TicketPrice,
		Asset:        sdk.AssetHive, // Default to HIVE, could be parameterized
		WinnerShares: args.WinnerShares,
		Pool:         0,
		Participants: make(map[string]uint64),
		State:        LotteryStateActive,
		Winners:      []Winner{},
		TotalTickets: 0,
	}

	// Save lottery
	saveLottery(lottery)

	// Emit event
	emitLotteryCreated(lottery)

	ret := "lottery created with ID: " + strconv.FormatUint(lottery.ID, 10)
	return &ret
}

//export join_lottery
func join_lottery(payload *string) *string {
	payloadStr := unwrapPayload(payload, "join_lottery payload missing")
	args := parseJoinLottery(payloadStr)

	// Require transfer intent
	transfer := getFirstTransferAllow()
	if transfer == nil {
		sdk.Abort("transfer.allow intent required")
	}

	sender := getSenderAddress()
	now := nowUnix()

	// Load only metadata (not participants) to check conditions
	meta := loadLotteryMetadata(args.LotteryID)
	if meta == nil {
		sdk.Abort("lottery not found")
	}

	// Check lottery is active
	if meta.State != LotteryStateActive {
		sdk.Abort("lottery is not active")
	}

	// Check deadline not passed
	if now >= meta.DeadlineUnix {
		sdk.Abort("lottery deadline has passed")
	}

	// Check asset matches
	if transfer.Token.String() != meta.Asset.String() {
		sdk.Abort("asset mismatch")
	}

	// Calculate how many tickets can be bought
	totalAmount := FloatToAmount(transfer.Limit)
	if totalAmount < meta.TicketPrice {
		sdk.Abort("insufficient funds for at least one ticket")
	}

	ticketCount := uint64(totalAmount / meta.TicketPrice)
	actualCost := Amount(ticketCount) * meta.TicketPrice

	// Draw funds from sender to contract
	sdk.HiveDraw(AmountToInt64(actualCost), meta.Asset)

	senderStr := sender.String()

	// Load pool stats
	stats := loadLotteryPoolStats(args.LotteryID)

	// Check if participant already exists
	participantIndex := loadParticipantIndex(args.LotteryID, senderStr)

	if participantIndex == 0 {
		// New participant - increment count and assign index
		stats.ParticipantCount++
		participantIndex = stats.ParticipantCount

		// Save participant entry
		entry := &ParticipantEntry{
			Address: senderStr,
			Tickets: ticketCount,
		}
		saveParticipantEntry(args.LotteryID, participantIndex, entry)

		// Save lookup index
		saveParticipantIndex(args.LotteryID, senderStr, participantIndex)
	} else {
		// Existing participant - load and update tickets
		entry := loadParticipantEntry(args.LotteryID, participantIndex)
		if entry != nil {
			entry.Tickets += ticketCount
			saveParticipantEntry(args.LotteryID, participantIndex, entry)
		}
	}

	// Update pool stats
	stats.Pool += actualCost
	stats.TotalTickets += ticketCount
	saveLotteryPoolStats(args.LotteryID, stats)

	// Emit event
	emitLotteryJoined(args.LotteryID, sender, ticketCount, actualCost, meta.Asset)

	ret := "joined lottery with " + strconv.FormatUint(ticketCount, 10) + " ticket(s)"
	return &ret
}

//export execute_lottery
func execute_lottery(payload *string) *string {
	payloadStr := unwrapPayload(payload, "execute_lottery payload missing")
	args := parseExecuteLottery(payloadStr)

	now := nowUnix()

	// Load lottery
	lottery := loadLottery(args.LotteryID)
	if lottery == nil {
		sdk.Abort("lottery not found")
	}

	// Check lottery is active
	if lottery.State != LotteryStateActive {
		sdk.Abort("lottery already executed")
	}

	// Check deadline has passed
	if now < lottery.DeadlineUnix {
		sdk.Abort("lottery deadline has not passed yet")
	}

	// Check there are participants
	if lottery.TotalTickets == 0 {
		sdk.Abort("no participants in lottery")
	}

	// Generate random seed
	lottery.RandomSeed = generateRandomSeed()

	// Calculate burn amount
	burnAmount := Amount(float64(lottery.Pool) * lottery.BurnPercent / 100.0)
	lottery.BurnedAmount = burnAmount

	// Burn tokens by sending to null
	nullReceiver := AddressFromString("hive:null")
	if burnAmount > 0 {
		sdk.HiveWithdraw(nullReceiver, AmountToInt64(burnAmount), lottery.Asset)
	}

	// Calculate remaining pool for distribution
	remainingPool := lottery.Pool - burnAmount

	// Select winners
	winnerCount := len(lottery.WinnerShares)
	winnerAddresses := selectRandomWinners(lottery.Participants, lottery.TotalTickets, winnerCount, lottery.RandomSeed)

	// Handle case where we have fewer participants than winner spots
	actualWinnerCount := len(winnerAddresses)

	// Distribute prizes
	lottery.Winners = make([]Winner, 0, actualWinnerCount)
	distributedTotal := Amount(0)

	for i, winnerAddr := range winnerAddresses {
		share := lottery.WinnerShares[i]
		winAmount := Amount(float64(remainingPool) * share / 100.0)

		if winAmount > 0 {
			sdk.HiveTransfer(winnerAddr, AmountToInt64(winAmount), lottery.Asset)
		}

		winner := Winner{
			Address: winnerAddr,
			Amount:  winAmount,
			Share:   share,
		}
		lottery.Winners = append(lottery.Winners, winner)
		distributedTotal += winAmount

		// Emit payout event
		emitLotteryPayout(lottery.ID, winnerAddr, winAmount, share, lottery.Asset, i+1)
	}

	// Send any undistributed funds to null (unclaimed shares + rounding remainder)
	if distributedTotal < remainingPool {
		undistributed := remainingPool - distributedTotal
		nullReceiver := AddressFromString("hive:null")
		sdk.HiveWithdraw(nullReceiver, AmountToInt64(undistributed), lottery.Asset)
		// Update total burned amount to include undistributed funds
		lottery.BurnedAmount += undistributed
	}

	// Update lottery state
	lottery.State = LotteryStateExecuted
	lottery.ExecutedAt = now

	// Save lottery
	saveLottery(lottery)

	// Emit execution event
	emitLotteryExecuted(lottery)

	ret := "lottery executed with " + strconv.FormatUint(uint64(len(lottery.Winners)), 10) + " winner(s)"
	return &ret
}

//export verify_lottery
func verify_lottery(payload *string) *string {
	payloadStr := unwrapPayload(payload, "verify_lottery payload missing")
	args := parseVerifyLottery(payloadStr)

	// Load lottery (read-only, no state changes)
	lottery := loadLottery(args.LotteryID)
	if lottery == nil {
		sdk.Abort("lottery not found")
	}

	// Lottery must be executed to verify
	if lottery.State != LotteryStateExecuted {
		sdk.Abort("lottery not executed yet - nothing to verify")
	}

	// Re-run winner selection with provided seed
	winnerCount := len(lottery.WinnerShares)
	verifiedWinners := selectRandomWinners(lottery.Participants, lottery.TotalTickets, winnerCount, args.Seed)

	// Compare with actual winners
	actualWinnerCount := len(lottery.Winners)
	verifiedWinnerCount := len(verifiedWinners)

	if actualWinnerCount != verifiedWinnerCount {
		ret := "verification failed: winner count mismatch"
		return &ret
	}

	// Check if all winners match
	allMatch := true
	for i := 0; i < actualWinnerCount; i++ {
		if lottery.Winners[i].Address.String() != verifiedWinners[i].String() {
			allMatch = false
			break
		}
	}

	if !allMatch {
		ret := "verification failed: winners do not match"
		return &ret
	}

	// Build result with winner list
	ret := "verification successful: " + strconv.FormatUint(uint64(actualWinnerCount), 10) + " winner(s) match"
	for i, winner := range lottery.Winners {
		ret += "|" + strconv.Itoa(i+1) + ":" + winner.Address.String()
	}

	return &ret
}
