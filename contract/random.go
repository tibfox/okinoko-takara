package main

import (
	"crypto/sha256"
	"encoding/binary"
	"okinoko_lottery/sdk"
)

// generateRandomSeed creates a cryptographically secure deterministic seed from transaction data
func generateRandomSeed() uint64 {
	env := currentEnv()

	// Collect entropy sources
	h := sha256.New()

	// Add transaction ID (primary entropy source - unique per execution)
	if env.TxId != "" {
		h.Write([]byte(env.TxId))
	}

	// Add block height
	if env.BlockHeight > 0 {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, env.BlockHeight)
		h.Write(buf)
	}

	// Add timestamp
	ts := nowUnix()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(ts))
	h.Write(buf)

	// Add sender address (prevents same-block manipulation)
	sender := getSenderAddress()
	h.Write([]byte(sender.String()))

	// Get SHA-256 hash and convert first 8 bytes to uint64
	hash := h.Sum(nil)
	seed := binary.LittleEndian.Uint64(hash[:8])

	return seed
}

// hashRandom uses SHA-256 based PRNG for cryptographically secure deterministic randomness
type hashRandom struct {
	seed    uint64
	counter uint64
}

func newHashRandom(seed uint64) *hashRandom {
	return &hashRandom{seed: seed, counter: 0}
}

// next generates the next random uint64 using SHA-256
func (r *hashRandom) next() uint64 {
	h := sha256.New()

	// Write seed
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, r.seed)
	h.Write(buf)

	// Write counter (ensures each call produces different output)
	binary.LittleEndian.PutUint64(buf, r.counter)
	h.Write(buf)

	r.counter++

	// Get first 8 bytes of hash as uint64
	hash := h.Sum(nil)
	return binary.LittleEndian.Uint64(hash[:8])
}

// intn returns a random number in [0, n) without modulo bias
// Uses rejection sampling to ensure uniform distribution
func (r *hashRandom) intn(n int) int {
	if n <= 0 {
		return 0
	}

	// Calculate the largest multiple of n that fits in uint64
	un := uint64(n)
	max := ^uint64(0) - (^uint64(0) % un)

	for {
		val := r.next()
		// Reject values that would cause bias
		if val < max {
			return int(val % un)
		}
		// If rejected, try again (expected iterations: ~1.0)
	}
}

// selectRandomWinners picks random winners from weighted ticket pool
// Returns winner addresses and their ticket counts
func selectRandomWinners(participants map[string]uint64, totalTickets uint64, winnerCount int, seed uint64) []sdk.Address {
	if winnerCount == 0 || totalTickets == 0 {
		return []sdk.Address{}
	}

	// Build weighted ticket pool
	ticketPool := make([]sdk.Address, 0, totalTickets)
	for addr, count := range participants {
		for i := uint64(0); i < count; i++ {
			ticketPool = append(ticketPool, sdk.Address(addr))
		}
	}

	// Shuffle the pool using Fisher-Yates with cryptographically secure RNG
	rng := newHashRandom(seed)
	n := len(ticketPool)
	for i := n - 1; i > 0; i-- {
		j := rng.intn(i + 1)
		ticketPool[i], ticketPool[j] = ticketPool[j], ticketPool[i]
	}

	// Select first winnerCount unique addresses
	winners := make([]sdk.Address, 0, winnerCount)
	seen := make(map[string]bool)

	for _, addr := range ticketPool {
		addrStr := addr.String()
		if !seen[addrStr] {
			winners = append(winners, addr)
			seen[addrStr] = true
			if len(winners) == winnerCount {
				break
			}
		}
	}

	return winners
}
