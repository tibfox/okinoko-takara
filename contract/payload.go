package main

import (
	"okinoko_lottery/sdk"
	"strconv"
	"strings"
)

// parseCreateLottery parses the payload for create_lottery
// Format: name|deadlineHours|burnPercent|winnerShare1,winnerShare2,...|ticketPrice[|donationAccount|donationPercent][|metaData][|max_tickets=<count>]
// Example: "My Lottery|168|10|50,30,20|5.000" or "My Lottery|168|10|50,30,20|5.000|hive:charity|5|My meta|max_tickets=1000"
func parseCreateLottery(payload string) *CreateLotteryArgs {
	parts := strings.Split(payload, "|")
	if len(parts) < 5 || len(parts) > 9 {
		sdk.Abort("invalid create_lottery payload format: expected 5 to 9 parts")
	}

	maxTickets := uint64(0)
	if len(parts) > 5 {
		last := strings.TrimSpace(parts[len(parts)-1])
		if strings.HasPrefix(last, "max_tickets=") {
			value := strings.TrimSpace(strings.TrimPrefix(last, "max_tickets="))
			if value == "" {
				sdk.Abort("max tickets must be greater than 0")
			}
			parsed, err := strconv.ParseUint(value, 10, 64)
			if err != nil || parsed == 0 {
				sdk.Abort("max tickets must be greater than 0")
			}
			maxTickets = parsed
			parts = parts[:len(parts)-1]
		}
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

	deadlineHours, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		sdk.Abort("invalid deadline hours")
	}
	if deadlineHours < 1 {
		sdk.Abort("deadline must be at least 1 hour")
	}
	if deadlineHours > 2160 {
		sdk.Abort("deadline must be 2160 hours or less")
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

	args := &CreateLotteryArgs{
		Name:            name,
		DeadlineHours:   deadlineHours,
		MaxTickets:      maxTickets,
		BurnPercent:     burnPercent,
		WinnerShares:    winnerShares,
		TicketPrice:     FloatToAmount(ticketPrice),
		DonationAccount: sdk.Address(""),
		DonationPercent: 0.0,
		MetaData:        "",
	}

	// Parse optional donation parameters
	if len(parts) == 7 || len(parts) == 8 {
		donationAccount := strings.TrimSpace(parts[5])
		if donationAccount == "" {
			sdk.Abort("donation account cannot be empty if provided")
		}
		// donation account could be a dao project in future - for now just basic user address
		args.DonationAccount = sdk.Address(donationAccount)

		donationPercent, err := strconv.ParseFloat(strings.TrimSpace(parts[6]), 64)
		if err != nil {
			sdk.Abort("invalid donation percent")
		}
		if donationPercent < 0.0 || donationPercent > 50.0 {
			sdk.Abort("donation percent must be between 0 and 50")
		}
		args.DonationPercent = donationPercent

		// Validate total percentages don't exceed 90% so that at least 10% goes to winners
		if burnPercent+donationPercent > 90.0 {
			sdk.Abort("burn percent + donation percent must not exceed 90")
		}
	}

	// Parse optional metadata
	if len(parts) == 6 {
		args.MetaData = strings.TrimSpace(parts[5])
	} else if len(parts) == 8 {
		args.MetaData = strings.TrimSpace(parts[7])
	} else {
		sdk.Abort("invalid create_lottery payload format: expected 6 or 8 parts for metadata")
	}
	if len(args.MetaData) > 500 {
		sdk.Abort("metadata must be 500 characters or less")
	}

	return args
}

// parseChangeLotteryMetadata parses the payload for change_lottery_metadata
// Format: lotteryID|metaData
// Example: "1|New metadata for the lottery"
func parseChangeLotteryMetadata(payload string) *ChangeLotteryMetadataArgs {
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		sdk.Abort("invalid change_lottery_metadata payload format: expected lotteryID|metaData")
	}

	lotteryID, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		sdk.Abort("invalid lottery ID")
	}
	if lotteryID == 0 {
		sdk.Abort("lottery ID must be greater than 0")
	}

	metaData := strings.TrimSpace(parts[1])
	if len(metaData) > 500 {
		sdk.Abort("metadata must be 500 characters or less")
	}

	return &ChangeLotteryMetadataArgs{
		LotteryID: lotteryID,
		MetaData:  metaData,
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

// parseVerifyLottery parses the payload for verify_lottery
// Format: lotteryID|seed
// Example: "1|12345678901234567890"
func parseVerifyLottery(payload string) *VerifyLotteryArgs {
	parts := strings.Split(payload, "|")
	if len(parts) != 2 {
		sdk.Abort("invalid verify_lottery payload format: expected lotteryID|seed")
	}

	lotteryID, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		sdk.Abort("invalid lottery ID")
	}
	if lotteryID == 0 {
		sdk.Abort("lottery ID must be greater than 0")
	}

	seed, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		sdk.Abort("invalid seed")
	}

	return &VerifyLotteryArgs{
		LotteryID: lotteryID,
		Seed:      seed,
	}
}
