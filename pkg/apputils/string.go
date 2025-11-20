package apputils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func StringToSnakeCase(s string) string {
	var result strings.Builder
	result.Grow(len(s) + 5) // Approximate additional space for underscores

	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result = make([]byte, length)
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))] // Pick random character from charset
	}

	return string(result)
}

func UUIDChecker(uuidString string) uuid.UUID {
	id, err := uuid.Parse(uuidString)
	if err != nil {
		log.Fatalf("Invalid UUID format: %v", err)
	}

	return id
}

func StringToInterface(payload string) interface{} {
	var result interface{}

	// Try to unmarshal as JSON
	if err := json.Unmarshal([]byte(payload), &result); err == nil {
		return result
	}

	// Try converting to int, float, or bool
	if intVal, err := strconv.Atoi(payload); err == nil {
		return intVal
	}
	if floatVal, err := strconv.ParseFloat(payload, 64); err == nil {
		return floatVal
	}
	if boolVal, err := strconv.ParseBool(payload); err == nil {
		return boolVal
	}

	// If all else fails, return the original string
	return payload
}

func DereferenceString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func FormatMoney(amount float64, currencySymbol string) string {
	// Format to 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)

	// Split into integer and decimal parts
	parts := strings.Split(formatted, ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	// Add thousands separators (dots) to integer part
	result := ""
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			result += "."
		}
		result += string(digit)
	}

	// Combine with currency symbol and decimal part
	return fmt.Sprintf("%s %s,%s", currencySymbol, result, decimalPart)
}

func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}
