# Lazy Jackpot Round Verification

A standalone Go script for verifying LazyBox jackpot rounds independently. This tool allows users to verify the fairness of completed jackpot rounds without trusting the server.

## What it verifies

✅ **Server Hash** - Confirms the server seed matches the pre-committed hash  
✅ **Client Seed** - Verifies the client seed generation from all bets  
✅ **Result Calculation** - Checks the provably fair random number generation  
✅ **Winner Selection** - Validates the winner based on calculated ranges  

## Usage

### 1. Get verification data

Make a POST request to your LazyBox instance:
```bash
curl "https://api.lazycoin.app/api/jackpot/verify?round_id={your_round_id_here}"
```

### 2. Save response to file
```bash
# Save the JSON response to a file
echo 'PASTE_JSON_RESPONSE_HERE' > round_data.json
```

### 3. Run verification
```bash
# Verify using file
go run verify_jackpot_round.go round_data.json

# Or verify using JSON string directly
go run verify_jackpot_round.go '{"success":true,"round_id":"..."}'
```

## Example Output

```
🎰 Verifying Jackpot Round #123 (64f8b2c...)
📊 Total Pot: 45.67 TON
🎯 Claimed Result: 42.156
🏆 Claimed Winner: EQA1...4B2C
============================================================
1️⃣  Verifying Server Hash...
    ✅ Server hash matches: a1b2c3d4e5f6...
2️⃣  Verifying Client Seed...
    ✅ Client seed matches: 9f8e7d6c5b4a...
3️⃣  Verifying Result Calculation...
    ✅ Result matches: 42.156
4️⃣  Verifying Winner Selection...
    ✅ Winner matches: EQA1...4B2C
5️⃣  Winner Ranges:
    🏆 EQA1...4B2C: 35.000 - 65.432 (30.4% chance, 13.75 TON)
       EQB2...5C3D: 0.000 - 35.000 (35.0% chance, 15.92 TON)
       EQC3...6D4E: 65.432 - 100.000 (34.6% chance, 15.00 TON)
    🎯 Result 42.156 falls in winner's range
============================================================
🎉 VERIFICATION PASSED! This round is provably fair.
```

## Requirements

- Go 1.19 or later
- No external dependencies (uses only standard library)

## Algorithm

The verification uses the same provably fair algorithm as the LazyBox server:

1. **Server Seed**: Pre-generated random seed (hash revealed before betting)
2. **Client Seed**: SHA-256 hash of all bets (player addresses + amounts + gift IDs)
3. **Result**: HMAC-SHA256(server_seed, combined_data) % 100001 / 1000.0
4. **Winner**: Player whose bet range contains the result value

This ensures complete transparency and verifiability of all jackpot rounds.
