package main

import (
	"fmt"
	"okinoko_lottery/sdk"
	"strconv"
	"strings"
)

// emitLotteryCreated logs a lottery creation event
func emitLotteryCreated(l *Lottery) {
	// Format: lc|id:<id>|creator:<creator>|name:<name>|created_at:<unix>|deadline:<unix>|burn:<percent>|ticket:<price>|asset:<asset>|winners:<count>|shares:<csv>|donation_account:<account>|donation_percent:<percent>

	var winnerShares strings.Builder
	for i, share := range l.WinnerShares {
		if i > 0 {
			winnerShares.WriteString(",")
		}
		winnerShares.WriteString(strconv.FormatFloat(share, 'f', 2, 64))
	}

	// Build event with optional donation info
	event := fmt.Sprintf(
		"lc|id:%d|creator:%s|name:%s|created_at:%d|deadline:%d|burn:%.2f|ticket:%.3f|asset:%s|winners:%d|shares:%s",
		l.ID,
		l.Creator.String(),
		l.Name,
		l.CreatedAt,
		l.DeadlineUnix,
		l.BurnPercent,
		AmountToFloat(l.TicketPrice),
		l.Asset.String(),
		len(l.WinnerShares),
		winnerShares.String(),
	)

	// Add donation info if configured
	if l.DonationPercent > 0.0 && l.DonationAccount.String() != "" {
		event += fmt.Sprintf("|donation_account:%s|donation_percent:%.2f", l.DonationAccount.String(), l.DonationPercent)
	}

	sdk.Log(event)
}

// emitLotteryJoined logs a lottery join event
func emitLotteryJoined(lotteryID uint64, participant sdk.Address, ticketCount uint64, totalPaid Amount, asset sdk.Asset, ticketStart uint64, ticketEnd uint64) {
	// Format: lj|id:<id>|participant:<address>|tickets:<count>|paid:<amount>|asset:<asset>|ticket_start:<start>|ticket_end:<end>

	event := fmt.Sprintf(
		"lj|id:%d|participant:%s|tickets:%d|paid:%.3f|asset:%s|ticket_start:%d|ticket_end:%d",
		lotteryID,
		participant.String(),
		ticketCount,
		AmountToFloat(totalPaid),
		asset.String(),
		ticketStart,
		ticketEnd,
	)

	sdk.Log(event)
}

// emitLotteryExecuted logs a lottery execution event
func emitLotteryExecuted(l *Lottery, participantCount uint64) {
	// Format: le|id:<id>|pool:%.3f|burned:%.3f|donated:%.3f|asset:<asset>|winners:<count>|seed:<seed>|tickets:<total>|participants:<count>|executed_at:<unix>

	event := fmt.Sprintf(
		"le|id:%d|pool:%.3f|burned:%.3f|donated:%.3f|asset:%s|winners:%d|seed:%d|tickets:%d|participants:%d|executed_at:%d",
		l.ID,
		AmountToFloat(l.Pool),
		AmountToFloat(l.BurnedAmount),
		AmountToFloat(l.DonatedAmount),
		l.Asset.String(),
		len(l.Winners),
		l.RandomSeed,
		l.TotalTickets,
		participantCount,
		l.ExecutedAt,
	)

	sdk.Log(event)
}

// emitLotteryPayout logs a winner payout event
func emitLotteryPayout(lotteryID uint64, winner sdk.Address, amount Amount, share float64, asset sdk.Asset, position int) {
	// Format: lp|id:<id>|winner:<address>|amount:<amount>|share:<percent>|asset:<asset>|position:<n>

	event := fmt.Sprintf(
		"lp|id:%d|winner:%s|amount:%.3f|share:%.2f|asset:%s|position:%d",
		lotteryID,
		winner.String(),
		AmountToFloat(amount),
		share,
		asset.String(),
		position,
	)

	sdk.Log(event)
}

// emitLotteryDonation logs a donation payout event
func emitLotteryDonation(lotteryID uint64, recipient sdk.Address, amount Amount, percent float64, asset sdk.Asset) {
	// Format: ld|id:<id>|recipient:<address>|amount:<amount>|percent:<percent>|asset:<asset>

	event := fmt.Sprintf(
		"ld|id:%d|recipient:%s|amount:%.3f|percent:%.2f|asset:%s",
		lotteryID,
		recipient.String(),
		AmountToFloat(amount),
		percent,
		asset.String(),
	)

	sdk.Log(event)
}

// emitLotteryUndistributed logs when undistributed funds are sent to null
func emitLotteryUndistributed(lotteryID uint64, amount Amount, asset sdk.Asset) {
	// Format: lu|id:<id>|amount:<amount>|asset:<asset>

	event := fmt.Sprintf(
		"lu|id:%d|amount:%.3f|asset:%s",
		lotteryID,
		AmountToFloat(amount),
		asset.String(),
	)

	sdk.Log(event)
}
