package main

import (
	"math"

	"okinoko_lottery/sdk"
)

const AmountScale = 1000

type Amount int64

// FloatToAmount scales human floats by AmountScale and rounds to int64 so storage stays precise.
func FloatToAmount(v float64) Amount {
	return Amount(math.Round(v * AmountScale))
}

// AmountToFloat converts back to float64 for reporting or events.
func AmountToFloat(v Amount) float64 {
	return float64(v) / AmountScale
}

// AmountToInt64 exposes the raw scaled int64 for Hive transfer functions.
func AmountToInt64(v Amount) int64 {
	return int64(v)
}

// LotteryState captures a lottery's lifecycle.
type LotteryState uint8

const (
	LotteryStateActive   LotteryState = 0
	LotteryStateExecuted LotteryState = 1
)

// String prints the lottery state as lower-case text for events and logs.
func (ls LotteryState) String() string {
	switch ls {
	case LotteryStateActive:
		return "active"
	case LotteryStateExecuted:
		return "executed"
	default:
		return "unknown"
	}
}

// Lottery represents a lottery instance
type Lottery struct {
	ID           uint64
	Creator      sdk.Address
	Name         string
	CreatedAt    int64
	DeadlineDays uint64
	DeadlineUnix int64
	BurnPercent  float64
	TicketPrice  Amount
	Asset        sdk.Asset
	WinnerShares []float64
	Pool         Amount
	Participants map[string]uint64 // address -> ticket count
	State        LotteryState
	Winners      []Winner
	ExecutedAt   int64
	RandomSeed   uint64
	TotalTickets uint64
	BurnedAmount Amount
}

// Winner represents a lottery winner
type Winner struct {
	Address sdk.Address
	Amount  Amount
	Share   float64
}

// CreateLotteryArgs represents arguments for creating a lottery
type CreateLotteryArgs struct {
	Name         string
	DeadlineDays uint64
	BurnPercent  float64
	WinnerShares []float64
	TicketPrice  Amount
	Asset        sdk.Asset
}

// JoinLotteryArgs represents arguments for joining a lottery
type JoinLotteryArgs struct {
	LotteryID uint64
}

// ExecuteLotteryArgs represents arguments for executing a lottery
type ExecuteLotteryArgs struct {
	LotteryID uint64
}

// AddressFromString converts a human string to the platform-specific address wrapper.
func AddressFromString(s string) sdk.Address { return sdk.Address(s) }

// AddressToString turns the wrapped type back into the underlying string.
func AddressToString(a sdk.Address) string { return a.String() }

// AssetFromString wraps a ticker string so type checking keeps us honest.
func AssetFromString(s string) sdk.Asset { return sdk.Asset(s) }

// AssetToString unwraps the ticker string for logs or SDK calls.
func AssetToString(a sdk.Asset) string { return a.String() }
