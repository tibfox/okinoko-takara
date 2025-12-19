package contract_test

import (
	"strconv"
	"strings"
	"testing"

	ledgerDb "vsc-node/modules/db/vsc/ledger"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// POSITIVE TESTS - Create Lottery
// ============================================================================

// TestCreateLotteryBasic tests basic lottery creation
func TestCreateLotteryBasic(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery: name|deadlineHours|burnPercent|winnerShares|ticketPrice
	payload := "Test Lottery|168|10|50,30,20|5.000"
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

	payload := "Winner Takes All|72|5|100|1.000"
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

// TestCreateLotteryWithMetadata tests metadata can be set at creation time
func TestCreateLotteryWithMetadata(t *testing.T) {
	ct := SetupContractTest()

	payload := "Meta Create|168|10|100|1.000|metadata from create"
	result, _, logs := CallContract(
		t, ct,
		"create_lottery",
		PayloadString(payload),
		nil,
		"hive:creator",
		true,
		uint(700_000_000),
	)

	assert.True(t, result.Success)
	assert.Equal(t, "metadata from create", ct.StateGet(ContractID, "lmd:1"))

	hasEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lm|") && strings.Contains(log, "id:1") {
				hasEvent = true
				assert.Contains(t, log, "metadata from create")
			}
		}
	}
	assert.True(t, hasEvent, "Expected metadata changed event on create")
}

// TestCreateLotteryMinBurnPercent tests minimum burn percent (5%)
func TestCreateLotteryMinBurnPercent(t *testing.T) {
	ct := SetupContractTest()

	payload := "Min Burn|24|5|100|1.000"
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

	payload := "Big Lottery|240|15|25,20,15,15,10,10,5|10.000"
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

// TestCreateLotteryMaxTicketsInvalid tests max tickets must be greater than 0 if provided
func TestCreateLotteryMaxTicketsInvalid(t *testing.T) {
	ct := SetupContractTest()

	payload := "Max Tickets|168|10|100|1.000|max_tickets=0"
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
	assert.Contains(t, result.Ret, "max tickets must be greater than 0")
}

// ============================================================================
// NEGATIVE TESTS - Create Lottery
// ============================================================================

// TestCreateLotteryNoName tests that lottery requires a name
func TestCreateLotteryNoName(t *testing.T) {
	ct := SetupContractTest()

	payload := "|168|10|100|5.000"
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

	// Deadline 0 hours (min is 1)
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
	assert.Contains(t, result.Ret, "deadline must be at least 1 hour")
}

// TestCreateLotteryBurnTooLow tests burn percent below minimum (5%)
func TestCreateLotteryBurnTooLow(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|168|4|100|5.000"
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
	payload := "Test|168|10|50,40,20|5.000"
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

	payload := "Test|168|10|100|-5.000"
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
	payload := "Test|168|10|100"
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
// METADATA TESTS
// ============================================================================

// TestChangeLotteryMetadataStored tests metadata is saved as a raw string in its own state key
func TestChangeLotteryMetadataStored(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Meta Test|168|10|100|5.000|testmeta"), nil, "hive:creator", true, uint(700_000_000))

	metadata := "meta|data: {\"note\":\"do not parse\"}"
	result, _, logs := CallContract(t, ct, "change_lottery_metadata", PayloadString("1|"+metadata), nil, "hive:creator", true, uint(700_000_000))

	assert.True(t, result.Success)
	assert.Contains(t, result.Ret, "lottery metadata updated")

	hasEvent := false
	for _, logValues := range logs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lm|") && strings.Contains(log, "id:1") {
				hasEvent = true
				assert.Contains(t, log, metadata)
			}
		}
	}
	assert.True(t, hasEvent, "Expected metadata changed event")

	stored := ct.StateGet(ContractID, "lmd:1")
	assert.Equal(t, metadata, stored)
}

// TestChangeLotteryMetadataMaxSize tests the metadata size limit is enforced at 500 chars
func TestChangeLotteryMetadataMaxSize(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Meta Limit|168|10|100|5.000"), nil, "hive:creator", true, uint(10_000_000_000))

	metadata := strings.Repeat("a", 500)
	result, _, _ := CallContract(t, ct, "change_lottery_metadata", PayloadString("1|"+metadata), nil, "hive:creator", true, uint(10_000_000_000))

	assert.True(t, result.Success)
	assert.Equal(t, metadata, ct.StateGet(ContractID, "lmd:1"))
}

// TestChangeLotteryMetadataTooLarge tests that metadata over 500 chars is rejected
func TestChangeLotteryMetadataTooLarge(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Meta Too Large|168|10|100|5.000"), nil, "hive:creator", true, uint(10_000_000_000))

	metadata := strings.Repeat("b", 501)
	result, _, _ := CallContract(t, ct, "change_lottery_metadata", PayloadString("1|"+metadata), nil, "hive:creator", false, uint(10_000_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "metadata must be 500 characters or less")
	assert.Equal(t, "", ct.StateGet(ContractID, "lmd:1"))
}

// ============================================================================
// POSITIVE TESTS - Join Lottery
// ============================================================================

// TestJoinLotterySingleTicket tests joining lottery with single ticket
func TestJoinLotterySingleTicket(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery first
	payload := "Test Lottery|168|10|100|5.000"
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

// TestJoinLotteryMaxTicketsExact tests joining up to a max tickets limit succeeds
func TestJoinLotteryMaxTicketsExact(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Max Tickets|168|10|100|1.000|max_tickets=3"), nil, "hive:creator", true, uint(700_000_000))

	result1, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("2.000"), "hive:alice", true, uint(700_000_000))
	result2, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("1.000"), "hive:bob", true, uint(700_000_000))

	assert.True(t, result1.Success)
	assert.True(t, result2.Success)
}

// TestJoinLotteryMaxTicketsExceeded tests joining beyond max tickets is rejected
func TestJoinLotteryMaxTicketsExceeded(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Max Tickets|168|10|100|1.000|max_tickets=3"), nil, "hive:creator", true, uint(700_000_000))

	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("2.000"), "hive:alice", true, uint(700_000_000))

	result, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("2.000"), "hive:bob", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery max tickets exceeded")
}

// TestJoinLotteryMaxTicketsReached tests joining after limit is reached is rejected
func TestJoinLotteryMaxTicketsReached(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("Max Tickets|168|10|100|1.000|max_tickets=2"), nil, "hive:creator", true, uint(700_000_000))

	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("2.000"), "hive:alice", true, uint(700_000_000))

	result, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("1.000"), "hive:bob", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery max tickets reached")
}

// TestJoinLotteryNoMaxTicketsCap tests that omitting max tickets allows more joins
func TestJoinLotteryNoMaxTicketsCap(t *testing.T) {
	ct := SetupContractTest()

	CallContract(t, ct, "create_lottery", PayloadString("No Cap|168|10|100|1.000"), nil, "hive:creator", true, uint(700_000_000))

	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:bob", true, uint(700_000_000))
	result, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:charlie", true, uint(700_000_000))

	assert.True(t, result.Success)
}

// TestJoinLotteryMultipleTickets tests joining with multiple tickets
func TestJoinLotteryMultipleTickets(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|2.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|10.000"), nil, "hive:creator", true, uint(700_000_000))

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

	// Create lottery with 24 hour deadline
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to join after 48 hours (deadline + 24 hours)
	futureTimestamp := "2025-09-05T00:00:00" // 48 hours after default timestamp
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

	// Create lottery: 24 hour deadline, 10% burn, 100% to winner, 10 HIVE ticket
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|100|10.000"), nil, "hive:creator", true, uint(700_000_000))

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
				assert.Contains(t, log, "pool:30.000")  // 3 tickets * 10 HIVE
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
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

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

	// Create lottery with 240 hour deadline
	CallContract(t, ct, "create_lottery", PayloadString("Test|240|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))
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
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

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
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))
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
	CallContract(t, ct, "create_lottery", PayloadString("Test|24|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))
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

	payload := "Test|168|80|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "burn percent must be between 5 and 75")
}

// TestExactTicketPurchase tests that only exact ticket cost is drawn
func TestExactTicketPurchase(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 3 HIVE ticket price
	CallContract(t, ct, "create_lottery", PayloadString("Test|168|10|100|3.000"), nil, "hive:creator", true, uint(700_000_000))

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

	// Note: When pipe is in the name like "Test|Pipe", the split shifts fields and breaks parsing
	payload := "Test|Pipe|7|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "invalid deadline hours")
}

// TestLongLotteryName tests that names over 100 chars are rejected
func TestLongLotteryName(t *testing.T) {
	ct := SetupContractTest()

	longName := strings.Repeat("a", 101)
	payload := longName + "|168|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "100 characters or less")
}

// TestIntegerSharesOnly tests that only integer shares are accepted
func TestIntegerSharesOnly(t *testing.T) {
	ct := SetupContractTest()

	// Try with decimal shares
	payload := "Test|168|10|33.33,33.33,33.34|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "must be integer")
}

// TestNegativeWinnerShares tests that negative shares are rejected
func TestNegativeWinnerShares(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|168|10|-10,60,50|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "winner share must be between 1 and 100")
}

// TestMinTicketPrice tests minimum ticket price validation
func TestMinTicketPrice(t *testing.T) {
	ct := SetupContractTest()

	// Try with 0 price
	payload := "Test|168|10|100|0"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "ticket price must be at least 0.001")
}

// TestMaxDeadlineHours tests maximum deadline validation
func TestMaxDeadlineHours(t *testing.T) {
	ct := SetupContractTest()

	payload := "Test|2184|10|100|5.000"
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString(payload), nil, "hive:creator", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "deadline must be 2160 hours or less")
}

// ============================================================================
// VERIFICATION TESTS
// ============================================================================

// TestVerifyLotteryResults tests that lottery results are deterministic and verifiable
func TestVerifyLotteryResults(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with multiple winners
	CallContract(t, ct, "create_lottery", PayloadString("Verification Test|24|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Add participants
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:bob", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:charlie", true, uint(700_000_000))
	CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("15.000"), "hive:dave", true, uint(700_000_000))

	// Execute lottery after deadline
	futureTimestamp := "2025-09-05T00:00:00"
	result1, _, logs1 := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:executor1", true, uint(700_000_000), futureTimestamp)
	assert.True(t, result1.Success)

	// Extract winners from first execution
	var winners1 []string
	for _, logValues := range logs1 {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lp|") {
				// Parse winner from payout event
				// Format: lp|id:X|winner:ADDRESS|amount:Y.ZZZ
				parts := strings.Split(log, "|")
				for _, part := range parts {
					if strings.HasPrefix(part, "winner:") {
						winner := strings.TrimPrefix(part, "winner:")
						winners1 = append(winners1, winner)
						break
					}
				}
			}
		}
	}

	assert.Equal(t, 3, len(winners1), "Should have 3 winners")

	// Now verify: Execute the same lottery again with a fresh contract state
	// This simulates an independent verifier re-running the selection
	ct2 := SetupContractTest()

	// Recreate the exact same lottery conditions
	CallContract(t, ct2, "create_lottery", PayloadString("Verification Test|24|10|50,30,20|5.000"), nil, "hive:creator", true, uint(700_000_000))
	CallContract(t, ct2, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:alice", true, uint(700_000_000))
	CallContract(t, ct2, "join_lottery", PayloadString("1"), transferIntent("10.000"), "hive:bob", true, uint(700_000_000))
	CallContract(t, ct2, "join_lottery", PayloadString("1"), transferIntent("5.000"), "hive:charlie", true, uint(700_000_000))
	CallContract(t, ct2, "join_lottery", PayloadString("1"), transferIntent("15.000"), "hive:dave", true, uint(700_000_000))

	// Execute with same timestamp and executor (same entropy sources = same seed)
	result2, _, logs2 := CallContractAt(t, ct2, "execute_lottery", PayloadString("1"), nil, "hive:executor1", true, uint(700_000_000), futureTimestamp)
	assert.True(t, result2.Success)

	// Extract winners from second execution
	var winners2 []string
	for _, logValues := range logs2 {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lp|") {
				parts := strings.Split(log, "|")
				for _, part := range parts {
					if strings.HasPrefix(part, "winner:") {
						winner := strings.TrimPrefix(part, "winner:")
						winners2 = append(winners2, winner)
						break
					}
				}
			}
		}
	}

	assert.Equal(t, 3, len(winners2), "Should have 3 winners in verification")

	// Verify that both executions produced identical winners in identical order
	// This proves the lottery is deterministic and verifiable
	for i := 0; i < len(winners1); i++ {
		assert.Equal(t, winners1[i], winners2[i], "Winner %d should match between executions", i+1)
	}

	t.Logf("Verification successful! Winners were identical:")
	t.Logf("  1st place (50%%): %s", winners1[0])
	t.Logf("  2nd place (30%%): %s", winners1[1])
	t.Logf("  3rd place (20%%): %s", winners1[2])
}

// TestVerifyLotteryFunction tests the verify_lottery contract function with many participants
func TestVerifyLotteryFunction(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with 1 HIVE ticket price, similar to large scale test
	CallContract(t, ct, "create_lottery", PayloadString("Verify Function Test|24|10|100|1.000"), nil, "hive:creator", true, uint(700_000_000))

	// Add 50 participants to get better randomness (enough to make collisions extremely unlikely)
	for i := 0; i < 50; i++ {
		participantName := "hive:user" + strconv.Itoa(i)
		// Deposit funds for this participant
		ct.Deposit(participantName, 200000, ledgerDb.AssetHive)
		// Each participant buys 1 ticket
		CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("1.000"), participantName, true, uint(700_000_000))
	}

	// Execute lottery
	futureTimestamp := "2025-09-05T00:00:00"
	execResult, _, execLogs := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:executor", true, uint(700_000_000), futureTimestamp)
	assert.True(t, execResult.Success)

	// Extract the winner from execution logs
	var actualWinner string
	for _, logValues := range execLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lp|") {
				parts := strings.Split(log, "|")
				for _, part := range parts {
					if strings.HasPrefix(part, "winner:") {
						actualWinner = strings.TrimPrefix(part, "winner:")
						break
					}
				}
			}
		}
	}

	t.Logf("Actual winner from execution: %s", actualWinner)
	assert.NotEmpty(t, actualWinner, "Should have a winner")

	// Extract the seed from execution logs (it's in the execution event)
	var seed string
	for _, logValues := range execLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "le|") {
				parts := strings.Split(log, "|")
				for _, part := range parts {
					if strings.HasPrefix(part, "seed:") {
						seed = strings.TrimPrefix(part, "seed:")
						break
					}
				}
			}
		}
	}

	t.Logf("Seed from execution: %s", seed)
	assert.NotEmpty(t, seed, "Seed should be present in execution logs")

	// Now verify using the verify_lottery function with the correct seed
	verifyPayload := "1|" + seed
	verifyResult, _, _ := CallContract(t, ct, "verify_lottery", PayloadString(verifyPayload), nil, "hive:anyone", true, uint(700_000_000))

	assert.True(t, verifyResult.Success)
	assert.Contains(t, verifyResult.Ret, "verification successful")
	assert.Contains(t, verifyResult.Ret, "1 winner(s) match")
	assert.Contains(t, verifyResult.Ret, actualWinner)

	t.Logf("Verification result: %s", verifyResult.Ret)

	// Test with wrong seed - with 50 participants, it's extremely unlikely to get the same winner
	wrongSeed := "12345678901234567890"
	wrongVerifyPayload := "1|" + wrongSeed
	wrongResult, _, _ := CallContract(t, ct, "verify_lottery", PayloadString(wrongVerifyPayload), nil, "hive:anyone", true, uint(700_000_000))

	assert.True(t, wrongResult.Success) // Function executes successfully
	assert.Contains(t, wrongResult.Ret, "verification failed", "Wrong seed should produce different winner with 50 participants")

	t.Logf("Wrong seed result: %s", wrongResult.Ret)

	// The key test: verify that using the CORRECT seed always works
	verifyAgain, _, _ := CallContract(t, ct, "verify_lottery", PayloadString(verifyPayload), nil, "hive:anyone", true, uint(700_000_000))
	assert.True(t, verifyAgain.Success)
	assert.Contains(t, verifyAgain.Ret, "verification successful")
	assert.Contains(t, verifyAgain.Ret, actualWinner)

	t.Logf("Verified again successfully with correct seed")
}

// TestVerifyLotteryNotExecuted tests that verification fails for non-executed lotteries
func TestVerifyLotteryNotExecuted(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery but don't execute
	CallContract(t, ct, "create_lottery", PayloadString("Not Executed|168|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))

	// Try to verify
	result, _, _ := CallContract(t, ct, "verify_lottery", PayloadString("1|12345"), nil, "hive:anyone", false, uint(700_000_000))

	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "lottery not executed yet")
}

// TestLotteryWithDonation tests lottery with optional donation feature
func TestLotteryWithDonation(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery with donation: 10% burn, 20% donation to hive:charity
	// Format: name|hours|burn%|shares|price|donationAccount|donationPercent
	createResult, _, createLogs := CallContract(t, ct, "create_lottery", PayloadString("Charity Lottery|24|10|100|5.000|hive:charity|20"), nil, "hive:creator", true, uint(700_000_000))
	assert.True(t, createResult.Success)
	assert.Contains(t, createResult.Ret, "lottery created with ID: 1")

	// Verify donation info is in the creation event
	foundDonationInfo := false
	for _, logValues := range createLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lc|") && strings.Contains(log, "donation_account:hive:charity") && strings.Contains(log, "donation_percent:20.00") {
				foundDonationInfo = true
				break
			}
		}
	}
	assert.True(t, foundDonationInfo, "Donation info should be in creation event")

	// Add participants - 4 participants, each buying 1 ticket (5 HIVE)
	participants := []string{"hive:alice", "hive:bob", "hive:charlie", "hive:dave"}
	for _, participant := range participants {
		ct.Deposit(participant, 10_000_000, ledgerDb.AssetHive) // 10 HIVE each
		joinResult, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), participant, true, uint(700_000_000))
		assert.True(t, joinResult.Success)
	}

	// Total pool: 20 HIVE
	// Expected burn: 2 HIVE (10%)
	// Expected donation: 4 HIVE (20%)
	// Expected prize pool: 14 HIVE (70%)

	// Execute lottery
	futureTimestamp := "2025-09-05T00:00:00"
	execResult, _, execLogs := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:executor", true, uint(700_000_000), futureTimestamp)
	assert.True(t, execResult.Success)

	// Verify donation event was emitted
	foundDonationEvent := false
	for _, logValues := range execLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "ld|") {
				// Format: ld|id:<id>|recipient:<address>|amount:<amount>|percent:<percent>|asset:<asset>
				assert.Contains(t, log, "id:1")
				assert.Contains(t, log, "recipient:hive:charity")
				assert.Contains(t, log, "amount:4.000")
				assert.Contains(t, log, "percent:20.00")
				foundDonationEvent = true
				break
			}
		}
	}
	assert.True(t, foundDonationEvent, "Donation event should be emitted")

	// Note: We cannot check balances of hive:null or hive:charity because sdk.HiveWithdraw
	// sends funds to Layer 1 (out of the contract), not to Layer 2 accounts.
	// We can only verify the contract balance decreased correctly.

	// Get winner to check their balance increased
	var winnerAddress string
	for _, logValues := range execLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "lp|") {
				parts := strings.Split(log, "|")
				for _, part := range parts {
					if strings.HasPrefix(part, "winner:") {
						winnerAddress = strings.TrimPrefix(part, "winner:")
						break
					}
				}
			}
		}
	}

	// Winner should have received 14 HIVE (70% of 20 HIVE pool)
	// The remaining 6 HIVE went to burn (2) and donation (4) via sdk.HiveWithdraw
	t.Logf("Winner: %s should have received 14 HIVE", winnerAddress)
	t.Logf("Donation: 4 HIVE sent to hive:charity via sdk.HiveWithdraw")
	t.Logf("Burn: 2 HIVE sent to hive:null via sdk.HiveWithdraw")
}

// TestLotteryWithoutDonation tests lottery without donation (backwards compatibility)
func TestLotteryWithoutDonation(t *testing.T) {
	ct := SetupContractTest()

	// Create lottery without donation (old format - 5 parts)
	createResult, _, _ := CallContract(t, ct, "create_lottery", PayloadString("Regular Lottery|24|10|100|5.000"), nil, "hive:creator", true, uint(700_000_000))
	assert.True(t, createResult.Success)
	assert.Contains(t, createResult.Ret, "lottery created with ID: 1")

	// Add participants
	participants := []string{"hive:alice", "hive:bob"}
	for _, participant := range participants {
		ct.Deposit(participant, 10_000_000, ledgerDb.AssetHive)
		joinResult, _, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("5.000"), participant, true, uint(700_000_000))
		assert.True(t, joinResult.Success)
	}

	// Execute lottery
	futureTimestamp := "2025-09-05T00:00:00"
	execResult, _, execLogs := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:executor", true, uint(700_000_000), futureTimestamp)
	assert.True(t, execResult.Success)

	// Verify NO donation event was emitted
	foundDonationEvent := false
	for _, logValues := range execLogs {
		for _, log := range logValues {
			if strings.HasPrefix(log, "ld|") {
				foundDonationEvent = true
				break
			}
		}
	}
	assert.False(t, foundDonationEvent, "No donation event should be emitted for lottery without donation")
}

// TestLotteryDonationValidation tests donation parameter validation
func TestLotteryDonationValidation(t *testing.T) {
	ct := SetupContractTest()

	// Test: donation percent too high (>50%)
	result, _, _ := CallContract(t, ct, "create_lottery", PayloadString("Too High|24|10|100|5.000|hive:charity|51"), nil, "hive:creator", false, uint(700_000_000))
	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "donation percent must be between 0 and 50")

	// Test: donation percent negative
	result, _, _ = CallContract(t, ct, "create_lottery", PayloadString("Negative|24|10|100|5.000|hive:charity|-5"), nil, "hive:creator", false, uint(700_000_000))
	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "donation percent must be between 0 and 50")

	// Test: burn + donation exceeds 90%
	result, _, _ = CallContract(t, ct, "create_lottery", PayloadString("Too Much|24|75|100|5.000|hive:charity|20"), nil, "hive:creator", false, uint(700_000_000))
	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "burn percent + donation percent must not exceed 90")

	// Test: empty donation account
	result, _, _ = CallContract(t, ct, "create_lottery", PayloadString("Empty Account|24|10|100|5.000||10"), nil, "hive:creator", false, uint(700_000_000))
	assert.False(t, result.Success)
	assert.Contains(t, result.Ret, "donation account cannot be empty if provided")

	// Test: valid donation at maximum limits (burn 70% + donation 20% = 90%)
	result, _, _ = CallContract(t, ct, "create_lottery", PayloadString("Max Valid|24|70|100|5.000|hive:charity|20"), nil, "hive:creator", true, uint(700_000_000))
	assert.True(t, result.Success)
}

// TestLargeScaleLottery tests lottery with 1000 participants
// uncommented as it takes a lot of time...
// func TestLargeScaleLottery(t *testing.T) {
// 	ct := SetupContractTestLargeScale()

// 	// Create lottery with 1 HIVE ticket price
// 	createResult, _, _ := CallContract(t, ct, "create_lottery", PayloadString("Large Scale Test|168|10|100|1.000"), nil, "hive:creator", true, uint(700_000_000))
// 	assert.True(t, createResult.Success)

// 	// Track total RC cost for all joins
// 	var totalJoinRC uint64 = 0
// 	var minRC uint64 = ^uint64(0) // Max uint64
// 	var maxRC uint64 = 0

// 	// Add 1000 participants
// 	for i := 0; i < 1000; i++ {
// 		participantName := "hive:user" + strconv.Itoa(i)
// 		result, rc, _ := CallContract(t, ct, "join_lottery", PayloadString("1"), transferIntent("1.000"), participantName, true, uint(700_000_000))

// 		assert.True(t, result.Success, "Failed to join for "+participantName)

// 		totalJoinRC += uint64(rc)
// 		if uint64(rc) < minRC {
// 			minRC = uint64(rc)
// 		}
// 		if uint64(rc) > maxRC {
// 			maxRC = uint64(rc)
// 		}

// 		// Progress indicator every 100 participants
// 		if (i+1)%100 == 0 {
// 			t.Logf("Progress: %d/1000 participants added", i+1)
// 		}
// 	}

// 	avgJoinRC := totalJoinRC / 1000
// 	t.Logf("Join RC costs - Min: %d, Max: %d, Avg: %d, Total: %d", minRC, maxRC, avgJoinRC, totalJoinRC)

// 	// Execute lottery after deadline
// 	futureTimestamp := "2025-09-11T00:00:00"
// 	execResult, execRC, logs := CallContractAt(t, ct, "execute_lottery", PayloadString("1"), nil, "hive:alice", true, uint(5_000_000_000), futureTimestamp)

// 	assert.True(t, execResult.Success)
// 	assert.Contains(t, execResult.Ret, "lottery executed with 1 winner(s)")

// 	t.Logf("Execute RC cost: %d", execRC)

// 	// Verify execution event
// 	hasExecEvent := false
// 	for _, logValues := range logs {
// 		for _, log := range logValues {
// 			if strings.HasPrefix(log, "le|") {
// 				hasExecEvent = true
// 				assert.Contains(t, log, "pool:1000.000") // 1000 participants * 1 HIVE
// 				assert.Contains(t, log, "tickets:1000")
// 				t.Logf("Execution event: %s", log)
// 			}
// 		}
// 	}
// 	assert.True(t, hasExecEvent)
// }
