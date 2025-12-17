# Ōkinoko Takara

A decentralized lottery platform on the [Magi Network](https://magi.eco/) where you can create and participate in customizable lotteries with transparent prize distributions and provably fair winner selection.

**Recommended:** Use the [Okinoko.io](https://okinoko.io) UI for the best user experience.

---

## What is Ōkinoko Takara?

Ōkinoko Takara lets you create custom lotteries with your own rules or join existing ones. Every lottery is:

- **Transparent** – All rules are public and stored on-chain
- **Provably Fair** – Winner selection is deterministic and verifiable
- **Decentralized** – No central authority controls the lottery
- **Customizable** – Choose ticket prices, deadlines, prize distributions, and burn rates

Each lottery benefits the HIVE price by burning a portion of the ticket prices.

---

## How It Works

### Creating a Lottery

As a lottery creator, you define:

1. **Lottery Name** – Give your lottery a memorable name (1 - 100 characters)
2. **Duration** – Set how many days until the lottery closes (1-90 days)
3. **Ticket Price** – Choose the price per ticket in HIVE (min. 0.001 HIVE)
4. **Burn Rate** – Percentage of the pool to burn (5-75%)
5. **Prize Distribution** – Define how prizes are split among winners

**Example:**
- Name: "Happy New Year"
- Duration: 7 days
- Ticket Price: 1 HIVE
- Burn Rate: 30%
- Prize Distribution: 1st place gets 100%

### Joining a Lottery

To join a lottery:

1. Find an active lottery you want to join on [Okinoko.io](https://okinoko.io)
2. Fill out the form and the amount of tickets you want to buy
3. The more tickets you have, the better your chances of winning

You can join the same lottery multiple times to increase your odds or simply buy mutiple tickets at once.

### How Winners Are Selected

When the lottery deadline passes, anyone can execute the lottery:

1. A portion of the prize pool is burned (sent to `hive:null`)
2. Winners are selected randomly based on ticket weight
3. Prizes are automatically distributed to winners on the Magi Network
4. If there are fewer participants than winner positions, unclaimed prizes are also burned

**Important:** The more tickets you have, the higher your chance of winning!

---

## Lottery Parameters

### Duration
- Minimum: 1 day
- Maximum: 90 days

### Burn Rate
- Minimum: 5%
- Maximum: 75%

### Prize Distribution
- You can have one or multiple winners
- Winner shares must add up to 100%
- Examples:
  - Single winner: `100%`
  - Three winners: `50%, 30%, 20%`
  - Five winners: `30%, 25%, 20%, 15%, 10%`

### Ticket Pricing
- Any positive amount in HIVE (e.g., 1.000, 5.000, 10.000)

---

## Example Scenarios

### Winner Takes All

**Setup:**
- Name: "Big Jackpot"
- Duration: 7 days
- Ticket Price: 5 HIVE
- Burn Rate: 10%
- Winners: 1 (gets 100%)

**Participants:**
- Alice buys 1 ticket (5 HIVE)
- Bob buys 2 tickets (10 HIVE)
- Charlie buys 1 ticket (5 HIVE)

**Total Pool:** 20 HIVE

**After Execution:**
- Burned: 2 HIVE (10%)
- Prize Pool: 18 HIVE
- Winner: Selected randomly (Bob has 50% chance, Alice and Charlie each have 25%)
- Winner receives: 18 HIVE

---

### Multiple Winners

**Setup:**
- Name: "Triple Threat"
- Duration: 3 days
- Ticket Price: 10 HIVE
- Burn Rate: 15%
- Winners: 3 (50%, 30%, 20%)

**Participants:**
- 7 people buy tickets totaling 100 HIVE

**After Execution:**
- Burned: 15 HIVE (15%)
- Prize Pool: 85 HIVE
- 1st Place: 42.5 HIVE (50%)
- 2nd Place: 25.5 HIVE (30%)
- 3rd Place: 17 HIVE (20%)

---

## Unclaimed Prizes

If there are fewer participants than winner positions, the unclaimed prize shares are automatically burned to `hive:null`.

**Example:**
- Lottery has 3 winner positions (50%, 30%, 20%)
- Only 2 people participate
- After execution:
  - 1st place gets 50% of the remaining pool
  - 2nd place gets 30% of the remaining pool
  - The unclaimed 20% is burned along with the configured burn rate

---

## Events & Transparency

All lottery activities can be recorded by off-chain indexers.

- **Lottery Created** – When a new lottery is created
- **Participant Joined** – Every time someone buys tickets
- **Lottery Executed** – When winners are selected and prizes distributed
- **Prize Payout** – Individual payout records for each winner

These events can be tracked and verified by anyone.

---

## Security & Fairness

### Provably Fair Randomness
Winner selection uses cryptographically secure randomness (SHA-256) based on:
- Transaction ID (unique to each execution)
- Block height and timestamp
- Executor's address

This makes the lottery:
- **Verifiable** – Anyone can verify the results were determined fairly
- **Unpredictable** – Results cannot be predicted before execution
- **Transparent** – All randomness sources are public on-chain
- **Auditable** – Complete execution history is available

### Verifying Results

After a lottery is executed, anyone can independently verify that the winners were selected fairly. There are two ways to verify:

#### Method 1: Using the verify_lottery Contract Function (Recommended)

The simplest way to verify is using the built-in `verify_lottery` contract function:

```javascript
// Call the verify_lottery contract function
// Format: lotteryID|seed
const result = await contract.call("verify_lottery", "1|12345678901234567890")

// Result will be:
// Success: "verification successful: 3 winner(s) match|1:hive:alice|2:hive:bob|3:hive:charlie"
// Failure: "verification failed: winners do not match"
```

To verify a lottery:
1. Retrieve the lottery seed from the execution event or on-chain data
2. Call `verify_lottery` with the lottery ID and seed
3. The contract re-runs the selection algorithm and compares results
4. Returns success or failure with the winner list

**Key benefits:**
- ✅ No need to implement the algorithm yourself
- ✅ Works directly on-chain with a single contract call
- ✅ Anyone can verify without downloading any data

#### Method 2: Off-Chain Verification

For advanced users who want to verify independently without calling the contract:

```go
// 1. Load executed lottery data from chain
lottery := getLotteryFromChain(lotteryID)

// 2. Verify the lottery was executed
if lottery.State != Executed {
    return error("lottery not executed")
}

// 3. Re-run winner selection with the stored seed
verifiedWinners := selectRandomWinners(
    lottery.Participants,  // Participant addresses and ticket counts
    lottery.TotalTickets,  // Total number of tickets
    len(lottery.Winners),  // Number of winners
    lottery.RandomSeed,    // Seed used during execution
)

// 4. Compare results
for i, winner := range lottery.Winners {
    if winner.Address != verifiedWinners[i] {
        return error("winner mismatch - lottery may be compromised!")
    }
}

// Results match - lottery was provably fair!
```

Because the random seed is stored on-chain and the selection algorithm is deterministic, verification always produces the same results. This makes cheating impossible without detection.

### Deadline Enforcement
- You cannot join a lottery after its deadline
- A lottery cannot be executed before its deadline
- These rules are enforced by the smart contract

### No Creator Advantage
Anyone can execute a lottery after its deadline - the creator has no special privileges.

---

## Getting Started

The easiest way to use Ōkinoko Takara is through the **[Okinoko.io](https://okinoko.io)** web interface, where you can:

- Browse active lotteries
- Create your own lottery with custom parameters
- Join lotteries with a simple interface
- Track your participation and winnings
- View lottery history and results

---

## Support

For questions or support, visit:
- **Website:** [Okinoko.io](https://okinoko.io)
- **Magi Network:** [magi.eco](https://magi.eco/)

---

## Contract Parameters Quick Reference

| Action | Format | Example |
|--------|--------|---------|
| Create Lottery | `name\|days\|burn%\|shares\|price` | `Weekly Draw\|7\|10\|100\|5.000` |
| Join Lottery | `lotteryID` | `1` |
| Execute Lottery | `lotteryID` | `1` |
| Verify Lottery | `lotteryID\|seed` | `1\|12345678901234567890` |

**Notes:**
- When joining, you must also provide a `transfer.allow` intent with the amount of HIVE you want to spend on tickets.
- When verifying, use the seed from the lottery execution event to independently verify the results.
