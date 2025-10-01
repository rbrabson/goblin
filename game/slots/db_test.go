package slots

import (
	"testing"
)

func TestGetPayoutAverages(t *testing.T) {
	// Test with a non-existent guild (should return zero values without error)
	guildID := "test-guild-nonexistent"

	averages, err := GetPayoutAverages(guildID)
	if err != nil {
		t.Errorf("GetPayoutAverages() error = %v, want nil", err)
		return
	}

	if averages == nil {
		t.Error("GetPayoutAverages() returned nil, want non-nil PayoutAverages")
		return
	}

	// For a non-existent guild, all values should be zero
	if averages.AverageTotalWins != 0 {
		t.Errorf("GetPayoutAverages().AverageTotalWins = %v, want 0", averages.AverageTotalWins)
	}

	if averages.AverageTotalLosses != 0 {
		t.Errorf("GetPayoutAverages().AverageTotalLosses = %v, want 0", averages.AverageTotalLosses)
	}

	if averages.TotalBet != 0 {
		t.Errorf("GetPayoutAverages().TotalBet = %v, want 0", averages.TotalBet)
	}

	if averages.TotalWon != 0 {
		t.Errorf("GetPayoutAverages().TotalWon = %v, want 0", averages.TotalWon)
	}
}

func TestPayoutAveragesStruct(t *testing.T) {
	// Test that the PayoutAverages struct can be created and fields are accessible
	averages := &PayoutAverages{
		AverageTotalWins: 10.5,
		TotalBet:         1000,
		AverageReturns:   120.0,
	}

	if averages.AverageTotalWins != 10.5 {
		t.Errorf("PayoutAverages.AverageTotalWins = %v, want 10.5", averages.AverageTotalWins)
	}

	if averages.TotalBet != 1000 {
		t.Errorf("PayoutAverages.TotalBet = %v, want 1000", averages.TotalBet)
	}

	if averages.AverageReturns != 120.0 {
		t.Errorf("PayoutAverages.AverageReturns = %v, want 120.0", averages.AverageReturns)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test getFloatFromResult helper function
	testData := map[string]interface{}{
		"float64_val": float64(123.45),
		"int64_val":   int64(100),
		"nil_val":     nil,
	}

	result := getFloatFromResult(testData, "float64_val")
	if result != 123.45 {
		t.Errorf("getFloatFromResult(float64) = %v, want 123.45", result)
	}

	result = getFloatFromResult(testData, "int64_val")
	if result != 100.0 {
		t.Errorf("getFloatFromResult(int64) = %v, want 100.0", result)
	}

	result = getFloatFromResult(testData, "nonexistent")
	if result != 0.0 {
		t.Errorf("getFloatFromResult(nonexistent) = %v, want 0.0", result)
	}

	result = getFloatFromResult(testData, "nil_val")
	if result != 0.0 {
		t.Errorf("getFloatFromResult(nil) = %v, want 0.0", result)
	}

	// Test getInt64FromResult helper function
	intResult := getInt64FromResult(testData, "int64_val")
	if intResult != 100 {
		t.Errorf("getInt64FromResult(int64) = %v, want 100", intResult)
	}

	intResult = getInt64FromResult(testData, "float64_val")
	if intResult != 123 {
		t.Errorf("getInt64FromResult(float64) = %v, want 123", intResult)
	}

	intResult = getInt64FromResult(testData, "nonexistent")
	if intResult != 0 {
		t.Errorf("getInt64FromResult(nonexistent) = %v, want 0", intResult)
	}
}
