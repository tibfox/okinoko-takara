package main

import (
	"fmt"
	"okinoko_lottery/sdk"
	"strconv"
	"strings"
)

// emitLotteryCreated logs a lottery creation event
func emitLotteryCreated(l *Lottery) {
	// Format: lc|id:<id>|creator:<creator>|name:<name>|deadline:<unix>|burn:<percent>|ticket:<price>|asset:<asset>|winners:<count>

	var winnerShares strings.Builder
	for i, share := range l.WinnerShares {
		if i > 0 {
			winnerShares.WriteString(",")
		}
		winnerShares.WriteString(strconv.FormatFloat(share, 'f', 2, 64))
	}

	event := fmt.Sprintf(
		"lc|id:%d|creator:%s|name:%s|deadline:%d|burn:%.2f|ticket:%.3f|asset:%s|winners:%d|shares:%s",
		l.ID,
		l.Creator.String(),
		l.Name,
		l.DeadlineUnix,
		l.BurnPercent,
		AmountToFloat(l.TicketPrice),
		l.Asset.String(),
		len(l.WinnerShares),
		winnerShares.String(),
	)

	sdk.Log(event)
}

// emitLotteryJoined logs a lottery join event
func emitLotteryJoined(lotteryID uint64, participant sdk.Address, ticketCount uint64, totalPaid Amount, asset sdk.Asset) {
	// Format: lj|id:<id>|participant:<address>|tickets:<count>|paid:<amount>|asset:<asset>

	event := fmt.Sprintf(
		"lj|id:%d|participant:%s|tickets:%d|paid:%.3f|asset:%s",
		lotteryID,
		participant.String(),
		ticketCount,
		AmountToFloat(totalPaid),
		asset.String(),
	)

	sdk.Log(event)
}

// emitLotteryExecuted logs a lottery execution event
func emitLotteryExecuted(l *Lottery) {
	// Format: le|id:<id>|pool:%.3f|burned:%.3f|asset:<asset>|winners:<count>|seed:<seed>

	event := fmt.Sprintf(
		"le|id:%d|pool:%.3f|burned:%.3f|asset:%s|winners:%d|seed:%d|tickets:%d",
		l.ID,
		AmountToFloat(l.Pool),
		AmountToFloat(l.BurnedAmount),
		l.Asset.String(),
		len(l.Winners),
		l.RandomSeed,
		l.TotalTickets,
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
