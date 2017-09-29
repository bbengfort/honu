package honu

import (
	"encoding/json"
	"math/rand"
	"os"
	"strings"
)

//===========================================================================
// Results Aggregation Helpers
//===========================================================================

// Helper function to append json data as a one line string to the end of a
// results file without deleting the previous contents in it.
func appendJSON(path string, val interface{}) error {
	// Open the file for appending, creating it if necessary
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Marshal the JSON in one line without indents
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// Append a newline to the data
	data = append(data, byte('\n'))

	// Append the data to the file
	_, err = f.Write(data)
	return err
}

//===========================================================================
// Key generator
//===========================================================================

// generateKey generates a random three letter key. If a prefix is specified
// then it generates a key with that prefix, otherwise it generates the key
// with random values.
func generateKey(prefix string) string {
	if prefix == "" {
		return randomConsonant() + randomVowel() + randomConsonant()
	}

	prefix = strings.ToUpper(prefix)
	return prefix + randomVowel() + randomConsonant()

}

var vowels = []string{"A", "E", "I", "O", "U", "Y"}
var consonants = []string{
	"B", "C", "D", "F", "G", "H", "J", "K", "L", "M",
	"N", "P", "Q", "R", "S", "T", "V", "W", "X", "Z",
}

// randomVowel returns a random english vowel.
func randomVowel() string {
	return vowels[rand.Intn(len(vowels))]
}

// randomConsonant returns a random english consonant.
func randomConsonant() string {
	return consonants[rand.Intn(len(consonants))]
}
