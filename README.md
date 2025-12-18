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
6. **Donation (Optional)** – Optionally dedicate a percentage to a charity or cause (0-50%)

**Example:**
- Name: "Happy New Year"
- Duration: 7 days
- Ticket Price: 1 HIVE
- Burn Rate: 30%
- Prize Distribution: 1st place gets 100%
- Donation: 5% to hive:charity (optional)

### Joining a Lottery

To join a lottery:

1. Find an active lottery you want to join on [Okinoko.io](https://okinoko.io)
2. Fill out the form and the amount of tickets you want to buy
3. The more tickets you have, the better your chances of winning

You can join the same lottery multiple times to increase your odds or simply buy mutiple tickets at once.

### How Winners Are Selected

When the lottery deadline passes, anyone can execute the lottery:

1. A portion of the prize pool is burned (sent to `hive:null`)
2. If configured, a donation is sent to the specified account
3. Winners are selected randomly based on ticket weight
4. Prizes are automatically distributed to winners on the Magi Network
5. If there are fewer participants than winner positions, unclaimed prizes are also burned

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

### Donation (Optional)
- Minimum: 0% (no donation)
- Maximum: 50%
- The donation account must be a valid Hive address
- Combined burn rate + donation rate cannot exceed 90%

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

### Charity Lottery with Donation

**Setup:**
- Name: "Help the Ocean"
- Duration: 14 days
- Ticket Price: 2 HIVE
- Burn Rate: 10%
- Donation: 20% to hive:oceanDAO
- Winners: 2 (60%, 40%)

**Participants:**
- 10 people buy tickets totaling 50 HIVE

**After Execution:**
- Burned: 5 HIVE (10%)
- Donated: 10 HIVE (20%) → sent to hive:oceanDAO
- Prize Pool: 35 HIVE
- 1st Place: 21 HIVE (60%)
- 2nd Place: 14 HIVE (40%)

This allows lottery creators to support charitable causes while still offering attractive prizes to participants!

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

All lottery activities emit events that can be indexed by off-chain systems. Each event contains complete information needed for indexing, verification, and analytics.

### Event Types

#### 1. Lottery Created (`lc`)
Emitted when a new lottery is created.

**Format:**
```
lc|id:<id>|creator:<address>|name:<name>|created_at:<unix_timestamp>|deadline:<unix_timestamp>|burn:<percent>|ticket:<price>|asset:<asset>|winners:<count>|shares:<csv>|donation_account:<account>|donation_percent:<percent>
```

**Fields:**
- `id` – Unique lottery ID
- `creator` – Address of lottery creator
- `name` – Lottery name
- `created_at` – Creation timestamp (Unix)
- `deadline` – Deadline timestamp (Unix)
- `burn` – Burn percentage (e.g., 10.00)
- `ticket` – Ticket price (e.g., 5.000)
- `asset` – Asset type (e.g., HIVE)
- `winners` – Number of winner positions
- `shares` – Prize distribution CSV (e.g., "50.00,30.00,20.00")
- `donation_account` – (Optional) Donation recipient address
- `donation_percent` – (Optional) Donation percentage

**Example:**
```
lc|id:1|creator:hive:alice|name:Weekly Draw|created_at:1703001600|deadline:1703606400|burn:10.00|ticket:5.000|asset:HIVE|winners:3|shares:50.00,30.00,20.00
```

#### 2. Lottery Joined (`lj`)
Emitted every time someone buys tickets.

**Format:**
```
lj|id:<id>|participant:<address>|tickets:<count>|paid:<amount>|asset:<asset>|ticket_start:<start>|ticket_end:<end>
```

**Fields:**
- `id` – Lottery ID
- `participant` – Buyer's address
- `tickets` – Number of tickets purchased
- `paid` – Total amount paid (e.g., 15.000)
- `asset` – Asset type
- `ticket_start` – First ticket number in range (0-indexed)
- `ticket_end` – Last ticket number in range (inclusive)

**Example:**
```
lj|id:1|participant:hive:bob|tickets:3|paid:15.000|asset:HIVE|ticket_start:0|ticket_end:2
lj|id:1|participant:hive:bob|tickets:2|paid:10.000|asset:HIVE|ticket_start:3|ticket_end:4
```

**Note:** The ticket range allows indexers to track exactly which ticket numbers belong to each participant.

#### 3. Lottery Executed (`le`)
Emitted when a lottery is executed and winners are selected.

**Format:**
```
le|id:<id>|pool:<amount>|burned:<amount>|donated:<amount>|asset:<asset>|winners:<count>|seed:<seed>|tickets:<total>|participants:<count>|executed_at:<unix_timestamp>
```

**Fields:**
- `id` – Lottery ID
- `pool` – Total pool before distribution
- `burned` – Total amount burned (includes undistributed funds)
- `donated` – Amount donated to charity (0 if none)
- `asset` – Asset type
- `winners` – Number of actual winners
- `seed` – Random seed used for selection
- `tickets` – Total tickets sold
- `participants` – Number of unique participants
- `executed_at` – Execution timestamp (Unix)

**Example:**
```
le|id:1|pool:100.000|burned:15.500|donated:0.000|asset:HIVE|winners:3|seed:12345678901234567890|tickets:20|participants:5|executed_at:1703606500
```

#### 4. Lottery Payout (`lp`)
Emitted for each winner when prizes are distributed.

**Format:**
```
lp|id:<id>|winner:<address>|amount:<amount>|share:<percent>|asset:<asset>|position:<n>
```

**Fields:**
- `id` – Lottery ID
- `winner` – Winner's address
- `amount` – Prize amount (e.g., 42.250)
- `share` – Prize share percentage (e.g., 50.00)
- `asset` – Asset type
- `position` – Winner position (1st, 2nd, 3rd, etc.)

**Example:**
```
lp|id:1|winner:hive:charlie|amount:42.250|share:50.00|asset:HIVE|position:1
lp|id:1|winner:hive:alice|amount:25.350|share:30.00|asset:HIVE|position:2
lp|id:1|winner:hive:bob|amount:16.900|share:20.00|asset:HIVE|position:3
```

#### 5. Lottery Donation (`ld`)
Emitted when a donation is sent to the configured charity/cause.

**Format:**
```
ld|id:<id>|recipient:<address>|amount:<amount>|percent:<percent>|asset:<asset>
```

**Fields:**
- `id` – Lottery ID
- `recipient` – Donation recipient address
- `amount` – Donation amount
- `percent` – Donation percentage
- `asset` – Asset type

**Example:**
```
ld|id:1|recipient:hive:oceanDAO|amount:10.000|percent:10.00|asset:HIVE
```

#### 6. Lottery Undistributed (`lu`)
Emitted when undistributed funds (from rounding or unclaimed shares) are burned.

**Format:**
```
lu|id:<id>|amount:<amount>|asset:<asset>
```

**Fields:**
- `id` – Lottery ID
- `amount` – Amount of undistributed funds burned
- `asset` – Asset type

**Example:**
```
lu|id:1|amount:0.500|asset:HIVE
```

**Note:** This ensures complete accounting transparency. The total burned = configured burn + undistributed funds.

### For Indexer Developers

These events provide **complete information** to:
- ✅ Track exact ticket ownership and ranges per participant
- ✅ Verify lottery fairness independently using the seed
- ✅ Provide full accounting (every token accounted for)
- ✅ Reconstruct complete lottery history
- ✅ Display user's specific ticket numbers
- ✅ Track all timestamps and state changes
- ✅ Monitor donation flows
- ✅ Account for rounding errors and undistributed funds

All events can be tracked, indexed, and verified by anyone monitoring the blockchain.

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

After a lottery is executed, anyone can independently verify that the winners were selected fairly. 

#### Using the verify_lottery Contract Function

The built-in way to verify is using the `verify_lottery` contract function:

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

## Contract Parameters Quick Reference

| Action | Format | Example |
|--------|--------|---------|
| Create Lottery | `name\|days\|burn%\|shares\|price[\|donationAccount\|donationPercent]` | `Weekly Draw\|7\|10\|100\|5.000` or `Charity Draw\|7\|10\|100\|5.000\|hive:charity\|10` |
| Join Lottery | `lotteryID` | `1` |
| Execute Lottery | `lotteryID` | `1` |
| Verify Lottery | `lotteryID\|seed` | `1\|12345678901234567890` |

**Notes:**
- When creating a lottery, the donation parameters are optional. If omitted, no donation is configured.
- When joining, you must also provide a `transfer.allow` intent with the amount of HIVE you want to spend on tickets.
- When verifying, use the seed from the lottery execution event to independently verify the results.
