package main

import (
	"encoding/binary"
	"math"
	"okinoko_lottery/sdk"
)

// LotteryMetadata contains the static/rarely-changing lottery data
type LotteryMetadata struct {
	ID              uint64
	Creator         sdk.Address
	Name            string
	CreatedAt       int64
	DeadlineDays    uint64
	DeadlineUnix    int64
	BurnPercent     float64
	TicketPrice     Amount
	Asset           sdk.Asset
	WinnerShares    []float64
	State           LotteryState
	Winners         []Winner
	ExecutedAt      int64
	RandomSeed      uint64
	BurnedAmount    Amount
	DonationAccount sdk.Address
	DonationPercent float64
	DonatedAmount   Amount
}

// LotteryPoolStats contains pool and ticket totals (frequently updated)
type LotteryPoolStats struct {
	Pool              Amount
	TotalTickets      uint64
	ParticipantCount  uint64 // Number of unique participants
}

// ParticipantEntry represents a single participant
type ParticipantEntry struct {
	Address string
	Tickets uint64
}

// encodeLotteryMetadata encodes the static lottery metadata
func encodeLotteryMetadata(m *LotteryMetadata) string {
	buf := make([]byte, 0, 256)

	buf = appendUint64(buf, m.ID)
	buf = appendString(buf, m.Creator.String())
	buf = appendString(buf, m.Name)
	buf = appendInt64(buf, m.CreatedAt)
	buf = appendUint64(buf, m.DeadlineDays)
	buf = appendInt64(buf, m.DeadlineUnix)
	buf = appendFloat64(buf, m.BurnPercent)
	buf = appendInt64(buf, int64(m.TicketPrice))
	buf = appendString(buf, m.Asset.String())

	// WinnerShares slice
	buf = appendUint64(buf, uint64(len(m.WinnerShares)))
	for _, share := range m.WinnerShares {
		buf = appendFloat64(buf, share)
	}

	buf = append(buf, byte(m.State))

	// Winners slice
	buf = appendUint64(buf, uint64(len(m.Winners)))
	for _, w := range m.Winners {
		buf = appendString(buf, w.Address.String())
		buf = appendInt64(buf, int64(w.Amount))
		buf = appendFloat64(buf, w.Share)
	}

	buf = appendInt64(buf, m.ExecutedAt)
	buf = appendUint64(buf, m.RandomSeed)
	buf = appendInt64(buf, int64(m.BurnedAmount))

	// Donation fields
	buf = appendString(buf, m.DonationAccount.String())
	buf = appendFloat64(buf, m.DonationPercent)
	buf = appendInt64(buf, int64(m.DonatedAmount))

	return string(buf)
}

// decodeLotteryMetadata decodes the static lottery metadata
func decodeLotteryMetadata(data string) *LotteryMetadata {
	buf := []byte(data)
	offset := 0

	m := &LotteryMetadata{}

	m.ID, offset = readUint64(buf, offset)
	creatorStr, off := readString(buf, offset)
	m.Creator = AddressFromString(creatorStr)
	offset = off
	m.Name, offset = readString(buf, offset)
	m.CreatedAt, offset = readInt64(buf, offset)
	m.DeadlineDays, offset = readUint64(buf, offset)
	m.DeadlineUnix, offset = readInt64(buf, offset)
	m.BurnPercent, offset = readFloat64(buf, offset)
	ticketPrice, off := readInt64(buf, offset)
	m.TicketPrice = Amount(ticketPrice)
	offset = off
	assetStr, off := readString(buf, offset)
	m.Asset = AssetFromString(assetStr)
	offset = off

	// WinnerShares slice
	sharesLen, off := readUint64(buf, offset)
	offset = off
	m.WinnerShares = make([]float64, sharesLen)
	for i := uint64(0); i < sharesLen; i++ {
		m.WinnerShares[i], offset = readFloat64(buf, offset)
	}

	if offset >= len(buf) {
		sdk.Abort("decode error: unexpected end of data")
	}
	m.State = LotteryState(buf[offset])
	offset++

	// Winners slice
	winnersLen, off := readUint64(buf, offset)
	offset = off
	m.Winners = make([]Winner, winnersLen)
	for i := uint64(0); i < winnersLen; i++ {
		addrStr, off := readString(buf, offset)
		offset = off
		amount, off := readInt64(buf, offset)
		offset = off
		share, off := readFloat64(buf, offset)
		offset = off
		m.Winners[i] = Winner{
			Address: AddressFromString(addrStr),
			Amount:  Amount(amount),
			Share:   share,
		}
	}

	m.ExecutedAt, offset = readInt64(buf, offset)
	m.RandomSeed, offset = readUint64(buf, offset)
	burnedAmount, off := readInt64(buf, offset)
	m.BurnedAmount = Amount(burnedAmount)
	offset = off

	// Donation fields
	donationAccountStr, off := readString(buf, offset)
	m.DonationAccount = AddressFromString(donationAccountStr)
	offset = off
	m.DonationPercent, offset = readFloat64(buf, offset)
	donatedAmount, off := readInt64(buf, offset)
	m.DonatedAmount = Amount(donatedAmount)

	return m
}

// encodeLotteryPoolStats encodes pool statistics
func encodeLotteryPoolStats(s *LotteryPoolStats) string {
	buf := make([]byte, 0, 24)
	buf = appendInt64(buf, int64(s.Pool))
	buf = appendUint64(buf, s.TotalTickets)
	buf = appendUint64(buf, s.ParticipantCount)
	return string(buf)
}

// decodeLotteryPoolStats decodes pool statistics
func decodeLotteryPoolStats(data string) *LotteryPoolStats {
	buf := []byte(data)
	offset := 0

	s := &LotteryPoolStats{}
	pool, off := readInt64(buf, offset)
	s.Pool = Amount(pool)
	offset = off
	s.TotalTickets, offset = readUint64(buf, offset)
	s.ParticipantCount, offset = readUint64(buf, offset)

	return s
}

// encodeParticipantEntry encodes a participant entry (address and tickets)
func encodeParticipantEntry(p *ParticipantEntry) string {
	buf := make([]byte, 0, 32)
	buf = appendString(buf, p.Address)
	buf = appendUint64(buf, p.Tickets)
	return string(buf)
}

// decodeParticipantEntry decodes a participant entry
func decodeParticipantEntry(data string) *ParticipantEntry {
	buf := []byte(data)
	offset := 0

	p := &ParticipantEntry{}
	p.Address, offset = readString(buf, offset)
	p.Tickets, offset = readUint64(buf, offset)

	return p
}

// Binary encoding helpers

func appendUint64(buf []byte, v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	return append(buf, b...)
}

func appendInt64(buf []byte, v int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return append(buf, b...)
}

func appendFloat64(buf []byte, v float64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	return append(buf, b...)
}

func appendString(buf []byte, s string) []byte {
	// Length-prefixed string
	buf = appendUint64(buf, uint64(len(s)))
	return append(buf, []byte(s)...)
}

func readUint64(buf []byte, offset int) (uint64, int) {
	if offset+8 > len(buf) {
		sdk.Abort("decode error: insufficient data for uint64")
	}
	v := binary.LittleEndian.Uint64(buf[offset : offset+8])
	return v, offset + 8
}

func readInt64(buf []byte, offset int) (int64, int) {
	v, off := readUint64(buf, offset)
	return int64(v), off
}

func readFloat64(buf []byte, offset int) (float64, int) {
	v, off := readUint64(buf, offset)
	return math.Float64frombits(v), off
}

func readString(buf []byte, offset int) (string, int) {
	length, off := readUint64(buf, offset)
	if off+int(length) > len(buf) {
		sdk.Abort("decode error: insufficient data for string")
	}
	s := string(buf[off : off+int(length)])
	return s, off + int(length)
}
