package main

import (
	"okinoko_lottery/sdk"
	"strconv"
	"strings"
)

// parseCreateLottery parses the payload for create_lottery
// Format: name|deadlineDays|burnPercent|winnerShare1,winnerShare2,...|ticketPrice
// Example: "My Lottery|7|10|50,30,20|5.000"
func parseCreateLottery(payload string) *CreateLotteryArgs {
	parts := strings.Split(payload, "|")
	if len(parts) != 5 {
		sdk.Abort("invalid create_lottery payload format: expected 5 parts")
	}

	name := strings.TrimSpace(parts[0])
	if name == "" {
		sdk.Abort("lottery name is required")
	}
	if len(name) > 100 {
		sdk.Abort("lottery name must be 100 characters or less")
	}
	if strings.Contains(name, "|") {
		sdk.Abort("lottery name cannot contain pipe character")
	}

	deadlineDays, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		sdk.Abort("invalid deadline days")
	}
	if deadlineDays < 1 {
		sdk.Abort("deadline must be at least 1 day")
	}
	if deadlineDays > 90 {
		sdk.Abort("deadline must be 90 days or less")
	}

	burnPercent, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		sdk.Abort("invalid burn percent")
	}
	if burnPercent < 5.0 || burnPercent > 75.0 {
		sdk.Abort("burn percent must be between 5 and 75")
	}

	// Parse winner shares (integers only)
	shareStrs := strings.Split(strings.TrimSpace(parts[3]), ",")
	if len(shareStrs) == 0 {
		sdk.Abort("at least one winner required")
	}

	winnerShares := make([]float64, len(shareStrs))
	totalShares := 0.0
	for i, s := range shareStrs {
		shareInt, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			sdk.Abort("invalid winner share: must be integer")
		}
		if shareInt <= 0 || shareInt > 100 {
			sdk.Abort("winner share must be between 1 and 100")
		}
		share := float64(shareInt)
		winnerShares[i] = share
		totalShares += share
	}

	// Check shares sum to 100
	if totalShares != 100.0 {
		sdk.Abort("winner shares must sum to 100")
	}

	ticketPrice, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
	if err != nil {
		sdk.Abort("invalid ticket price")
	}
	if ticketPrice < 0.001 {
		sdk.Abort("ticket price must be at least 0.001")
	}

	return &CreateLotteryArgs{
		Name:         name,
		DeadlineDays: deadlineDays,
		BurnPercent:  burnPercent,
		WinnerShares: winnerShares,
		TicketPrice:  FloatToAmount(ticketPrice),
	}
}

// parseJoinLottery parses the payload for join_lottery
// Format: lotteryID
// Example: "1"
func parseJoinLottery(payload string) *JoinLotteryArgs {
	lotteryID, err := strconv.ParseUint(strings.TrimSpace(payload), 10, 64)
	if err != nil {
		sdk.Abort("invalid lottery ID")
	}
	if lotteryID == 0 {
		sdk.Abort("lottery ID must be greater than 0")
	}

	return &JoinLotteryArgs{
		LotteryID: lotteryID,
	}
}

// parseExecuteLottery parses the payload for execute_lottery
// Format: lotteryID
// Example: "1"
func parseExecuteLottery(payload string) *ExecuteLotteryArgs {
	lotteryID, err := strconv.ParseUint(strings.TrimSpace(payload), 10, 64)
	if err != nil {
		sdk.Abort("invalid lottery ID")
	}
	if lotteryID == 0 {
		sdk.Abort("lottery ID must be greater than 0")
	}

	return &ExecuteLotteryArgs{
		LotteryID: lotteryID,
	}
}
