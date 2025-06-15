package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strings"
)

// VerificationBet represents a bet for verification
type VerificationBet struct {
	PlayerAddress string  `json:"player_address"`
	Amount        float64 `json:"amount"`
	GiftID        string  `json:"gift_id"`
}

// RoundVerificationData contains all data needed for verification
type RoundVerificationData struct {
	Success       bool              `json:"success"`
	RoundID       string            `json:"round_id"`
	RoundNumber   int               `json:"round_number"`
	ServerSeed    string            `json:"server_seed"`
	ServerHash    string            `json:"server_hash"`
	ClientSeed    string            `json:"client_seed"`
	PreviousHash  string            `json:"previous_hash"`
	Bets          []VerificationBet `json:"bets"`
	Result        float64           `json:"result"`
	WinnerAddress string            `json:"winner_address"`
	TotalPot      float64           `json:"total_pot"`
	Error         string            `json:"error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run verify_jackpot_round.go <verification_data.json>")
		fmt.Println("OR: go run verify_jackpot_round.go '<json_string>'")
		fmt.Println("\nTo get verification data, make a POST request to /api/jackpot/verify with:")
		fmt.Println(`{"round_id": "your_round_id"}`)
		os.Exit(1)
	}

	var data RoundVerificationData
	input := os.Args[1]

	// Try to read as file first
	if fileData, err := os.ReadFile(input); err == nil {
		if err := json.Unmarshal(fileData, &data); err != nil {
			log.Fatalf("Failed to parse JSON from file: %v", err)
		}
	} else {
		// Try to parse as JSON string
		if err := json.Unmarshal([]byte(input), &data); err != nil {
			log.Fatalf("Failed to parse JSON string: %v", err)
		}
	}

	if !data.Success {
		log.Fatalf("Verification data contains error: %s", data.Error)
	}

	fmt.Printf("🎰 Verifying Jackpot Round #%d (%s)\n", data.RoundNumber, data.RoundID)
	fmt.Printf("📊 Total Pot: %.2f TON\n", data.TotalPot)
	fmt.Printf("🎯 Claimed Result: %.3f\n", data.Result)
	fmt.Printf("🏆 Claimed Winner: %s\n", data.WinnerAddress)
	fmt.Println(strings.Repeat("=", 60))

	passed := true

	fmt.Println("1️⃣  Verifying Server Hash...")
	expectedHash := hashString(data.ServerSeed)
	if expectedHash == data.ServerHash {
		fmt.Printf("    ✅ Server hash matches: %s\n", data.ServerHash[:16]+"...")
	} else {
		fmt.Printf("    ❌ Server hash mismatch!\n")
		fmt.Printf("       Expected: %s\n", expectedHash)
		fmt.Printf("       Got:      %s\n", data.ServerHash)
		passed = false
	}

	fmt.Println("2️⃣  Verifying Client Seed...")
	calculatedClientSeed := generateClientSeed(data.Bets)
	if calculatedClientSeed == data.ClientSeed {
		fmt.Printf("    ✅ Client seed matches: %s\n", data.ClientSeed[:16]+"...")
	} else {
		fmt.Printf("    ❌ Client seed mismatch!\n")
		fmt.Printf("       Calculated: %s\n", calculatedClientSeed)
		fmt.Printf("       Claimed:    %s\n", data.ClientSeed)
		passed = false
	}

	fmt.Println("3️⃣  Verifying Result Calculation...")
	calculatedResult := calculateResult(data.ServerSeed, data.ClientSeed, data.RoundNumber, data.PreviousHash)
	if fmt.Sprintf("%.3f", calculatedResult) == fmt.Sprintf("%.3f", data.Result) {
		fmt.Printf("    ✅ Result matches: %.3f\n", data.Result)
	} else {
		fmt.Printf("    ❌ Result mismatch!\n")
		fmt.Printf("       Calculated: %.3f\n", calculatedResult)
		fmt.Printf("       Claimed:    %.3f\n", data.Result)
		passed = false
	}

	fmt.Println("4️⃣  Verifying Winner Selection...")
	calculatedWinner := selectWinner(data.Bets, data.Result)
	if calculatedWinner == data.WinnerAddress {
		fmt.Printf("    ✅ Winner matches: %s\n", data.WinnerAddress)
	} else {
		fmt.Printf("    ❌ Winner mismatch!\n")
		fmt.Printf("       Calculated: %s\n", calculatedWinner)
		fmt.Printf("       Claimed:    %s\n", data.WinnerAddress)
		passed = false
	}

	fmt.Println("5️⃣  Winner Ranges:")
	showWinnerRanges(data.Bets, data.Result)

	fmt.Println(strings.Repeat("=", 60))
	if passed {
		fmt.Println("🎉 VERIFICATION PASSED! This round is provably fair.")
	} else {
		fmt.Println("💀 VERIFICATION FAILED! This round may not be fair.")
		os.Exit(1)
	}
}

func hashString(str string) string {
	h := sha256.Sum256([]byte(str))
	return hex.EncodeToString(h[:])
}

func generateClientSeed(bets []VerificationBet) string {
	// Sort bets by player address alphabetically
	sortedBets := make([]VerificationBet, len(bets))
	copy(sortedBets, bets)
	sort.Slice(sortedBets, func(i, j int) bool {
		return sortedBets[i].PlayerAddress < sortedBets[j].PlayerAddress
	})

	h := sha256.New()
	for _, bet := range sortedBets {
		h.Write([]byte(bet.PlayerAddress))
		h.Write([]byte(fmt.Sprintf("%.3f", bet.Amount)))
		h.Write([]byte(bet.GiftID))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func calculateResult(serverSeed, clientSeed string, roundNumber int, previousHash string) float64 {
	combined := fmt.Sprintf("%s:%s:%d:%s", serverSeed, clientSeed, roundNumber, previousHash)
	h := hmac.New(sha256.New, []byte(serverSeed))
	h.Write([]byte(combined))
	hash := h.Sum(nil)

	hashInt := new(big.Int).SetBytes(hash)
	maxValue := big.NewInt(100001)
	resultInt := new(big.Int).Mod(hashInt, maxValue)
	return float64(resultInt.Int64()) / 1000.0
}

func selectWinner(bets []VerificationBet, result float64) string {
	if len(bets) == 0 {
		return ""
	}

	// Sort bets by player address alphabetically
	sortedBets := make([]VerificationBet, len(bets))
	copy(sortedBets, bets)
	sort.Slice(sortedBets, func(i, j int) bool {
		return sortedBets[i].PlayerAddress < sortedBets[j].PlayerAddress
	})

	// Calculate total bet amount
	totalBets := 0.0
	for _, bet := range sortedBets {
		totalBets += bet.Amount
	}

	// Find winner based on result position
	currentPosition := 0.0
	for _, bet := range sortedBets {
		betPercentage := (bet.Amount / totalBets) * 100.0
		rangeEnd := currentPosition + betPercentage

		if result >= currentPosition && result < rangeEnd {
			return bet.PlayerAddress
		}

		currentPosition = rangeEnd
	}

	// Should never reach here if bets are valid
	return sortedBets[len(sortedBets)-1].PlayerAddress
}

func showWinnerRanges(bets []VerificationBet, result float64) {
	if len(bets) == 0 {
		fmt.Println("    No bets to show")
		return
	}

	// Sort bets by player address alphabetically
	sortedBets := make([]VerificationBet, len(bets))
	copy(sortedBets, bets)
	sort.Slice(sortedBets, func(i, j int) bool {
		return sortedBets[i].PlayerAddress < sortedBets[j].PlayerAddress
	})

	// Calculate total bet amount
	totalBets := 0.0
	for _, bet := range sortedBets {
		totalBets += bet.Amount
	}

	// Show ranges
	currentPosition := 0.0
	for _, bet := range sortedBets {
		betPercentage := (bet.Amount / totalBets) * 100.0
		rangeEnd := currentPosition + betPercentage

		isWinner := result >= currentPosition && result < rangeEnd
		winnerIcon := "  "
		if isWinner {
			winnerIcon = "🏆"
		}

		playerDisplay := bet.PlayerAddress
		if len(playerDisplay) > 8 {
			playerDisplay = playerDisplay[:4] + "..." + playerDisplay[len(playerDisplay)-4:]
		}

		fmt.Printf("    %s %s: %.3f - %.3f (%.1f%% chance, %.2f TON)\n",
			winnerIcon, playerDisplay, currentPosition, rangeEnd, betPercentage, bet.Amount)

		currentPosition = rangeEnd
	}

	fmt.Printf("    🎯 Result %.3f falls in winner's range\n", result)
}
