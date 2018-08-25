package smpp

import "github.com/donovansolms/go-smpp/smpp/pdu/pdutext"

// GetMessageCount determines the amount of messages needed for the given
// text
func GetMessageCount(message pdutext.Codec) int {

	maxLength := 160
	messageParts := 1

	rawMsg := message.Encode()
	messageLength := len(rawMsg)
	if messageLength > maxLength {
		// Now check the amount of messages
		maxLength := 153 // 160-7 (UDH with 2 byte reference number) (bytes)
		if message.Type() == pdutext.UCS2Type {
			maxLength = 132 // to avoid a character being split between payloads  (bytes)
		}
		messageParts = int((len(rawMsg)-1)/maxLength) + 1

	}
	return messageParts
}

// ContainsSpecialCharacters checks if the given input contains non-ascii/latin1
// characters
func ContainsSpecialCharacters(input string) bool {
	for _, char := range input {
		if char > 127 {
			return true
		}
	}
	return false
}