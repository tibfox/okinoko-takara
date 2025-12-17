package contract_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// POSITIVE TESTS - Create Lottery
// ============================================================================

// TestCreateLotteryBasic tests basic lottery creation
func TestCreateLotteryBasic(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery: name|deadlineDays|burnPercent|winnerShares|ticketPrice
	payload := "Test Lottery|7|10|50,30,20|5.000"
	result, _, logs := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil, // No intent needed for creation
		"hive:creator",
		true,
		uint(700_000_000), // Increased for WASM JSON marshaling
	)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "lottery created with ID: 1")

	// Check for lottery created event
	hasEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lc|") && strings.Contains(log, "id:1") {
				hasEvent = true
				assert.Contains(t, log, "name:Test Lottery")
				assert.Contains(t, log, "burn:10.00")
				assert.Contains(t, log, "ticket:5.000")
				assert.Contains(t, log, "winners:3")
			}
		}
	}
	assert.True(t, hasEvent, "Expected lottery created event")
}

// TestCreateLotterySingleWinner tests lottery with single winner
func TestCreateLotterySingleWinner(t *testing.T) {
	ct := SetupContractTest()

	payload := "Winner Takes All|3|5|100|1.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
}

// TestCreateLotteryMinBurnPercent tests minimum burn percent (5%)
func TestCreateLotteryMinBurnPercent(t *testing.T) {
	ct := SetupContractTest()

	payload := "Min Burn|1|5|100|1.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
}

// TestCreateLotteryMultipleWinners tests lottery with many winners
func TestCreateLotteryMultipleWinners(t *testing.T) {
	ct := SetupContractTest()

	payload := "Big Lottery|10|15|25,20,15,15,10,10,5|10.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
}

// ============================================================================
// NEGATIVE TESTS - Create Lottery
// ============================================================================

// TestCreateLotteryNoName tests that lottery requires a name
func TestCreateLotteryNoName(t *testing.T) {
	ct := SetupContractTest()

	payload := "|7|10|100|5.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "name is required")
}

// TestCreateLotteryInvalidDeadline tests deadline validation
func TestCreateLotteryInvalidDeadline(t *testing.T) {
	ct := SetupContractTest()

	// Deadline 0 days (min is 1)
	payload := "Test|0|10|100|5.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "deadline must be at least 1 day")
}

// TestCreateLotteryBurnTooLow tests burn percent below minimum (5%)
func TestCreateLotteryBurnTooLow(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|7|4|100|5.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "burn percent must be between 5 and 75")
}

// TestCreateLotterySharesNotSum100 tests that shares must sum to 100
func TestCreateLotterySharesNotSum100(t *testing.T) {
	ct := SetupContractTest()

	// Shares sum to 110
	payload := "Test|7|10|50,40,20|5.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "shares must sum to 100")
}

// TestCreateLotteryNegativeTicketPrice tests that ticket price must be positive
func TestCreateLotteryNegativeTicketPrice(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|7|10|100|-5.000"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
}

// TestCreateLotteryInvalidFormat tests payload format validation
func TestCreateLotteryInvalidFormat(t *testing.T) {
	ct := SetupContractTest()

	// Missing parts
	payload := "Test|7|10|100"
	result, _, _ := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "invalid create_lottery payload format")
}

// ============================================================================
// POSITIVE TESTS - Join Lottery
// ============================================================================

// TestJoinLotterySingleTicket tests joining lottery with single ticket
func TestJoinLotterySingleTicket(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery first
	payload := "Test Lottery|7|10|100|5.000"
	CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", true, uint(700_000_000))

	// Join with 1 ticket (5 HIVE)
	result, _, logs := CallContract(
		t, ct,
		"join_lottery",
		PayloadString("1"),
		transferIntent("5.000"),
		"hive:alice",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "joined lottery with 1 ticket(s)")

	// Check for join event
	hasEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lj|") {
				hasEvent = true
				assert.Contains(t, log, "tickets:1")
				assert.Contains(t, log, "participant:hive:alice")
			}
		}
	}
	assert.True(t, hasEvent)
}

// TestJoinLotteryMultipleTickets tests joining with multiple tickets
func TestJoinLotteryMultipleTickets(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Join with 3 tickets (15 HIVE)
	result, _, logs := CallContract(
		t, ct,
		"join_lottery",
		PayloadString("1"),
		transferIntent("15.000"),
		"hive:bob",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "joined lottery with 3 ticket(s)")

	// Verify correct ticket count
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lj|") {
				assert.Contains(t, log, "tickets:3")
			}
		}
	}
}

// TestJoinLotteryMultipleParticipants tests multiple people joining
func TestJoinLotteryMultipleParticipants(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|2.000"), nil, "hive:creator", true, uint(700_000_000))

	// Multiple people join
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("2.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("4.000"), "hive:bob", true, uint(700_000_000))
	result, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("6.000"), "hive:charlie", true, uint(700_000_000))

	assert.True(t, result.Success)
}

// TestJoinLotterySamePersonMultipleTimes tests joining multiple times
func TestJoinLotterySamePersonMultipleTimes(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Same person joins twice
	result1, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	result2, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:alice", true, uint(700_000_000))

	assert.True(t, result1.Success)
	assert.True(t, result2.Success)
	assert.Contains(t, result2.Ret, "joined lottery with 2 ticket(s)")
}

// ============================================================================
// NEGATIVE TESTS - Join Lottery
// ============================================================================

// TestJoinLotteryNoIntent tests that join requires transfer intent
func TestJoinLotteryNoIntent(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to join without intent
	result, _, _ := CallContract(
		t, ct,
		"join_lottery",
		PayloadString("1"),
		nil, // No intent
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "transfer.allow intent required")
}

// TestJoinLotteryNotFound tests joining non-existent lottery
func TestJoinLotteryNotFound(t *testing.T) {
	ct := SetupContractTest()

	result, _, _ := CallContract(
		t, ct,
		"join_lottery",
		PayloadString("999"),
		transferIntent("5.000"),
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery not found")
}

// TestJoinLotteryInsufficientFunds tests joining with insufficient funds
func TestJoinLotteryInsufficientFunds(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 10 HIVE ticket price
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|10.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to join with only 5 HIVE
	result, _, _ := CallContract(
		t, ct,
		"join_lottery",
		PayloadString("1"),
		transferIntent("5.000"),
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "insufficient funds for at least one ticket")
}

// TestJoinLotteryAfterDeadline tests that joining after deadline fails
func TestJoinLotteryAfterDeadline(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 1 day deadline
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to join after 2 days (deadline + 1 day)
	futureTimestamp := "2025-09-05T00:00:00" // 2 days after default timestamp
	result, _, _ := CallContractAt(
		t, ct,
		"join_lottery",
		PayloadString("1"),
		transferIntent("5.000"),
		"hive:alice",
		false,
		uint(700_000_000),
		futureTimestamp,
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "deadline has passed")
}

// ============================================================================
// POSITIVE TESTS - Execute Lottery
// ============================================================================

// TestExecuteLotterySingleWinner tests executing lottery with one winner
func TestExecuteLotterySingleWinner(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery: 1 day deadline, 10% burn, 100% to winner, 10 HIVE ticket
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|100|10.000"), nil, "hive:creator", true, uint(700_000_000))

	// Three people join
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:bob", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:charlie", true, uint(700_000_000))

	// Execute after deadline
	futureTimestamp := "2025-09-05T00:00:00"
	result, _, logs := CallContractAt(
		t, ct,
		"execute_lottery",
		PayloadString("1"),
		nil,
		"hive:alice",
		true,
		uint(700_000_000),
		futureTimestamp,
	)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "lottery executed with 1 winner(s)")

	// Check execution event
	hasExecEvent := false
	hasPayoutEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "le|") {
				hasExecEvent = true
				assert.Contains(t, log, "pool:30.000") // 3 tickets * 10 HIVE
				assert.Contains(t, log, "burned:3.000") // 10% burn
				assert.Contains(t, log, "winners:1")
			}
			if strings.HasPrefix(log, "lp|") {
				hasPayoutEvent = true
				assert.Contains(t, log, "amount:27.000") // 30 - 3 burn
			}
		}
	}
	assert.True(t, hasExecEvent)
	assert.True(t, hasPayoutEvent)
}

// TestExecuteLotteryMultipleWinners tests lottery with multiple winners
func TestExecuteLotteryMultipleWinners(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery: 50%, 30%, 20% distribution
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Four people join
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:bob", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:charlie", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:dave", true, uint(700_000_000))

	// Execute
	futureTimestamp := "2025-09-05T00:00:00"
	result, _, logs := CallContractAt(
		t, ct,
		"execute_lottery",
		PayloadString("1"),
		nil,
		"hive:alice",
		true,
		uint(700_000_000),
		futureTimestamp,
	)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "lottery executed with 3 winner(s)")

	// Count payout events (should be 3)
	payoutCount := 0
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lp|") {
				payoutCount++
			}
		}
	}
	assert.Equal(t, 3, payoutCount)
}

// TestExecuteLotteryWithMultipleTickets tests weighted random selection
func TestExecuteLotteryWithMultipleTickets(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Alice buys 10 tickets, Bob buys 1 (Alice has 10x chance)
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("50.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:bob", true, uint(700_000_000))

	// Execute
	futureTimestamp := "2025-09-05T00:00:00"
	result, _, logs := CallContractAt(
		t, ct,
		"execute_lottery",
		PayloadString("1"),
		nil,
		"hive:alice",
		true,
		uint(700_000_000),
		futureTimestamp,
	)

	assert.True(t, result.Success)

	// Pool should be 55 HIVE, burned 5.5, distributed 49.5
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "le|") {
				assert.Contains(t, log, "pool:55.000")
				assert.Contains(t, log, "tickets:11")
			}
		}
	}
}

// ============================================================================
// NEGATIVE TESTS - Execute Lottery
// ============================================================================

// TestExecuteLotteryNotFound tests executing non-existent lottery
func TestExecuteLotteryNotFound(t *testing.T) {
	ct := SetupContractTest()

	result, _, _ := CallContract(
		t, ct,
		"execute_lottery",
		PayloadString("999"),
		nil,
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery not found")
}

// TestExecuteLotteryBeforeDeadline tests that execution before deadline fails
func TestExecuteLotteryBeforeDeadline(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 10 day deadline
	CallContract(t, ct, "create_lottery", PayloadString("Test|10|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))

	// Try to execute immediately (before deadline)
	result, _, _ := CallContract(
		t, ct,
		"execute_lottery",
		PayloadString("1"),
		nil,
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "deadline has not passed yet")
}

// TestExecuteLotteryNoParticipants tests executing lottery with no participants
func TestExecuteLotteryNoParticipants(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery but nobody joins
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to execute after deadline
	futureTimestamp := "2025-09-05T00:00:00"
	result, _, _ := CallContractAt(
		t, ct,
		"execute_lottery",
		PayloadString("1"),
		nil,
		"hive:alice",
		false,
		uint(700_000_000),
		futureTimestamp,
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "no participants")
}

// TestExecuteLotteryTwice tests that lottery can only be executed once
func TestExecuteLotteryTwice(t *testing.T) {
	ct := SetupContractTest()

	// Create and join lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))

	// Execute once
	futureTimestamp := "2025-09-05T00:00:00"
	result1, _, _ := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:alice", true, uint(700_000_000), futureTimestamp)
	assert.True(t, result1.Success)

	// Try to execute again
	result2, _, _ := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:alice", false, uint(700_000_000), futureTimestamp)
	assert.False(t, result2.Success)
	assert.Contains(t, result2.Ret, "already executed")
}

// TestExecuteLotteryInvalidID tests invalid lottery ID
func TestExecuteLotteryInvalidID(t *testing.T) {
	ct := SetupContractTest()

	result, _, _ := CallContract(
		t, ct,
		"execute_lottery",
		PayloadString("0"),
		nil,
		"hive:alice",
		false,
		uint(700_000_000),
	)

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery ID must be greater than 0")
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

// TestFewerParticipantsThanWinners tests lottery with fewer participants than winners
func TestFewerParticipantsThanWinners(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 3 winners but only 2 people join
	CallContract(t, ct, "create_lottery", PayloadString("Test|1|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:bob", true, uint(700_000_000))

	// Execute after deadline
	futureTimestamp := "2025-09-05T00:00:00"
	result, _, _ := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:alice", true, uint(700_000_000), futureTimestamp)

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "lottery executed with 2 winner(s)")
}

// TestMaxBurnPercent tests that burn percent is capped at 75%
func TestMaxBurnPercent(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|7|80|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "burn percent must be between 5 and 75")
}

// TestExactTicketPurchase tests that only exact ticket cost is drawn
func TestExactTicketPurchase(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 3 HIVE ticket price
	CallContract(t, ct, "create_lottery", PayloadString("Test|7|10|100|3.000"), nil, "hive:creator", true, uint(700_000_000))

	// Alice sends 10 HIVE intent but should only be charged 9 HIVE (3 tickets)
	result, _, logs := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:alice", true, uint(700_000_000))

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "joined lottery with 3 ticket(s)")

	// Verify only 9 HIVE was drawn in the logs
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lj|") {
				assert.Contains(t, log, "paid:9.000")
			}
		}
	}
}

// TestEmptyPayload tests that empty payloads are rejected
func TestEmptyPayload(t *testing.T) {
	ct := SetupContractTest()

	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(""), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	// Empty string gets split into array with 1 empty element, caught by format validation
	assert.Contains(t, result.Ret, "invalid create_lottery payload format")
}

// TestPipeInLotteryName tests that pipe characters in name are rejected
func TestPipeInLotteryName(t *testing.T) {
	ct := SetupContractTest()

	// Note: When pipe is in the name like "Test|Pipe", the split results in 6 parts instead of 5
	// So it gets caught by the format validation before name validation
	// The pipe character validation in code prevents edge cases where encoding might bypass this
	payload := "Test|Pipe|7|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	// Gets caught by format check because split creates 6 parts, not 5
	assert.Contains(t, result.Ret, "invalid create_lottery payload format")
}

// TestLongLotteryName tests that names over 100 chars are rejected
func TestLongLotteryName(t *testing.T) {
	ct := SetupContractTest()

	longName := strings.Repeat("a", 101)
	payload := longName + "|7|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "100 characters or less")
}

// TestIntegerSharesOnly tests that only integer shares are accepted
func TestIntegerSharesOnly(t *testing.T) {
	ct := SetupContractTest()

	// Try with decimal shares
	payload := "Test|7|10|33.33,33.33,33.34|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "must be integer")
}

// TestNegativeWinnerShares tests that negative shares are rejected
func TestNegativeWinnerShares(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|7|10|-10,60,50|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "winner share must be between 1 and 100")
}

// TestMinTicketPrice tests minimum ticket price validation
func TestMinTicketPrice(t *testing.T) {
	ct := SetupContractTest()

	// Try with 0 price
	payload := "Test|7|10|100|0"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "ticket price must be at least 0.001")
}

// TestMaxDeadlineDays tests maximum deadline validation
func TestMaxDeadlineDays(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|91|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "deadline must be 90 days or less")
}

// TestLargeScaleLottery tests lottery with 1000 participants
func TestLargeScaleLottery(t *testing.T) {
	ct := SetupContractTestLargeScale()

	// Create lottery with 1 HIVE ticket price
	createResult, _, _ := CallContract(t, ct, "create_lottery", PayloadString("Large Scale Test|7|10|100|1.000"), nil, "hive:creator", true, uint(700_000_000))
	assert.True(t, createResult.Success)

	// Track total RC cost for all joins
	var totalJoinRC uint64 = 0
	var minRC uint64 = ^uint64(0) // Max uint64
	var maxRC uint64 = 0

	// Add 1000 participants
	for i := 0; i < 1000; i++ {
		participantName := "hive:user" + strconv.Itoa(i)
		result, rc, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("1.000"), participantName, true, uint(700_000_000))

		assert.True(t, result.Success, "Failed to join for "+participantName)

		totalJoinRC += uint64(rc)
		if uint64(rc) < minRC {
			minRC = uint64(rc)
		}
		if uint64(rc) > maxRC {
			maxRC = uint64(rc)
		}

		// Progress indicator every 100 participants
		if (i+1)%100 == 0 {
			t.Logf("Progress: %d/1000 participants added", i+1)
		}
	}

	avgJoinRC := totalJoinRC / 1000
	t.Logf("Join RC costs - Min: %d, Max: %d, Avg: %d, Total: %d", minRC, maxRC, avgJoinRC, totalJoinRC)

	// Execute lottery after deadline
	futureTimestamp := "2025-09-11T00:00:00"
	execResult, execRC, logs := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:alice", true, uint(5_000_000_000), futureTimestamp)

	assert.True(t, execResult.Success)
	assert.Contains(t, execResult.Ret, "lottery executed with 1 winner(s)")

	t.Logf("Execute RC cost: %d", execRC)

	// Verify execution event
	hasExecEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "le|") {
				hasExecEvent = true
				assert.Contains(t, log, "pool:1000.000") // 1000 participants * 1 HIVE
				assert.Contains(t, log, "tickets:1000")
				t.Logf("Execution event: %s", log)
			}
		}
	}
	assert.True(t, hasExecEvent)
}
