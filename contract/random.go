package main

import (
	"okinoko_lottery/sdk"
)

// generateRandomSeed creates a deterministic seed from transaction data
func generateRandomSeed() uint64 {
	env := currentEnv()

	// Use transaction ID and block data to create a seed
	seed := uint64(0)

	// Hash the transaction ID
	if env.TxId != "" {
		for i := 0; i < len(env.TxId); i++ {
			seed = seed*31 + uint64(env.TxId[i])
		}
	}

	// Mix in block height if available
	if env.BlockHeight > 0 {
		seed ^= env.BlockHeight
	}

	// Mix in timestamp
	ts := nowUnix()
	seed ^= uint64(ts)

	return seed
}

// lcgRandom is a simple Linear Congruential Generator for deterministic randomness
type lcgRandom struct {
	state uint64
}

func newLCGRandom(seed uint64) *lcgRandom {
	return &lcgRandom{state: seed}
}

// next generates the next random number
func (r *lcgRandom) next() uint64 {
	// LCG parameters from Numerical Recipes
	r.state = (r.state*1664525 + 1013904223) & 0xFFFFFFFF
	return r.state
}

// intn returns a random number in [0, n)
func (r *lcgRandom) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next() % uint64(n))
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

	// Shuffle the pool using Fisher-Yates
	rng := newLCGRandom(seed)
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
