package main

import (
	"okinoko_lottery/sdk"
	"strconv"
)

// Gas-optimized state key builders
// Using compact string concatenation for minimal allocations

// getLotteryMetadataKey returns the storage key for lottery metadata by ID
func getLotteryMetadataKey(id uint64) string {
	return "lm:" + strconv.FormatUint(id, 10)
}

// getLotteryPoolStatsKey returns the storage key for lottery pool statistics
func getLotteryPoolStatsKey(id uint64) string {
	return "ls:" + strconv.FormatUint(id, 10)
}

// getParticipantIndexKey returns the storage key for a participant by index
func getParticipantIndexKey(lotteryID uint64, index uint64) string {
	return "lpi:" + strconv.FormatUint(lotteryID, 10) + ":" + strconv.FormatUint(index, 10)
}

// getParticipantLookupKey returns the storage key for looking up a participant's index by address
func getParticipantLookupKey(lotteryID uint64, address string) string {
	return "lpu:" + strconv.FormatUint(lotteryID, 10) + ":" + address
}

// getCounterKey returns the storage key for the lottery counter
func getCounterKey() string {
	return "counter"
}

// loadLotteryMetadata retrieves lottery metadata from state
func loadLotteryMetadata(id uint64) *LotteryMetadata {
	key := getLotteryMetadataKey(id)
	dataPtr := sdk.StateGetObject(key)
	if dataPtr == nil || *dataPtr == "" {
		return nil
	}
	return decodeLotteryMetadata(*dataPtr)
}

// saveLotteryMetadata stores lottery metadata to state
func saveLotteryMetadata(m *LotteryMetadata) {
	key := getLotteryMetadataKey(m.ID)
	data := encodeLotteryMetadata(m)
	sdk.StateSetObject(key, data)
}

// loadLotteryPoolStats retrieves lottery pool statistics
func loadLotteryPoolStats(id uint64) *LotteryPoolStats {
	key := getLotteryPoolStatsKey(id)
	dataPtr := sdk.StateGetObject(key)
	if dataPtr == nil || *dataPtr == "" {
		return &LotteryPoolStats{
			Pool:             0,
			TotalTickets:     0,
			ParticipantCount: 0,
		}
	}
	return decodeLotteryPoolStats(*dataPtr)
}

// saveLotteryPoolStats stores lottery pool statistics
func saveLotteryPoolStats(id uint64, s *LotteryPoolStats) {
	key := getLotteryPoolStatsKey(id)
	data := encodeLotteryPoolStats(s)
	sdk.StateSetObject(key, data)
}

// loadParticipantIndex retrieves a participant's index, returns 0 if not found
func loadParticipantIndex(lotteryID uint64, address string) uint64 {
	key := getParticipantLookupKey(lotteryID, address)
	dataPtr := sdk.StateGetObject(key)
	if dataPtr == nil || *dataPtr == "" {
		return 0
	}
	index, err := strconv.ParseUint(*dataPtr, 10, 64)
	if err != nil {
		sdk.Abort("invalid participant index")
	}
	return index
}

// saveParticipantIndex stores a participant's index
func saveParticipantIndex(lotteryID uint64, address string, index uint64) {
	key := getParticipantLookupKey(lotteryID, address)
	sdk.StateSetObject(key, strconv.FormatUint(index, 10))
}

// loadParticipantEntry retrieves a participant entry by index
func loadParticipantEntry(lotteryID uint64, index uint64) *ParticipantEntry {
	key := getParticipantIndexKey(lotteryID, index)
	dataPtr := sdk.StateGetObject(key)
	if dataPtr == nil || *dataPtr == "" {
		return nil
	}
	return decodeParticipantEntry(*dataPtr)
}

// saveParticipantEntry stores a participant entry at the given index
func saveParticipantEntry(lotteryID uint64, index uint64, entry *ParticipantEntry) {
	key := getParticipantIndexKey(lotteryID, index)
	data := encodeParticipantEntry(entry)
	sdk.StateSetObject(key, data)
}

// loadAllParticipants retrieves all participants for a lottery
func loadAllParticipants(lotteryID uint64) map[string]uint64 {
	stats := loadLotteryPoolStats(lotteryID)
	participants := make(map[string]uint64, stats.ParticipantCount)

	for i := uint64(1); i <= stats.ParticipantCount; i++ {
		entry := loadParticipantEntry(lotteryID, i)
		if entry != nil {
			participants[entry.Address] = entry.Tickets
		}
	}

	return participants
}

// loadLottery retrieves a full lottery from state (loads both metadata and participants)
func loadLottery(id uint64) *Lottery {
	meta := loadLotteryMetadata(id)
	if meta == nil {
		return nil
	}

	stats := loadLotteryPoolStats(id)
	participants := loadAllParticipants(id)

	// Combine into full lottery struct
	return &Lottery{
		ID:           meta.ID,
		Creator:      meta.Creator,
		Name:         meta.Name,
		CreatedAt:    meta.CreatedAt,
		DeadlineDays: meta.DeadlineDays,
		DeadlineUnix: meta.DeadlineUnix,
		BurnPercent:  meta.BurnPercent,
		TicketPrice:  meta.TicketPrice,
		Asset:        meta.Asset,
		WinnerShares: meta.WinnerShares,
		Pool:         stats.Pool,
		Participants: participants,
		State:        meta.State,
		Winners:      meta.Winners,
		ExecutedAt:   meta.ExecutedAt,
		RandomSeed:   meta.RandomSeed,
		TotalTickets: stats.TotalTickets,
		BurnedAmount: meta.BurnedAmount,
	}
}

// saveLottery stores a full lottery to state (saves both metadata and pool stats)
// Note: Individual participants are saved separately during join_lottery
func saveLottery(l *Lottery) {
	// Save metadata
	meta := &LotteryMetadata{
		ID:           l.ID,
		Creator:      l.Creator,
		Name:         l.Name,
		CreatedAt:    l.CreatedAt,
		DeadlineDays: l.DeadlineDays,
		DeadlineUnix: l.DeadlineUnix,
		BurnPercent:  l.BurnPercent,
		TicketPrice:  l.TicketPrice,
		Asset:        l.Asset,
		WinnerShares: l.WinnerShares,
		State:        l.State,
		Winners:      l.Winners,
		ExecutedAt:   l.ExecutedAt,
		RandomSeed:   l.RandomSeed,
		BurnedAmount: l.BurnedAmount,
	}
	saveLotteryMetadata(meta)

	// Save pool stats
	stats := &LotteryPoolStats{
		Pool:             l.Pool,
		TotalTickets:     l.TotalTickets,
		ParticipantCount: uint64(len(l.Participants)),
	}
	saveLotteryPoolStats(l.ID, stats)
}

// getNextLotteryID returns the next available lottery ID and increments the counter
func getNextLotteryID() uint64 {
	key := getCounterKey()
	counterPtr := sdk.StateGetObject(key)
	var counter uint64
	if counterPtr != nil && *counterPtr != "" {
		var err error
		counter, err = strconv.ParseUint(*counterPtr, 10, 64)
		if err != nil {
			sdk.Abort("invalid counter state")
		}
	}
	counter++
	sdk.StateSetObject(key, strconv.FormatUint(counter, 10))
	return counter
}
