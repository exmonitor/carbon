package sms

import "fmt"

func Send(number string, message string) error {
	// TODO
	fmt.Printf("<< Fake sms notification sent to %s\n", number)
	return nil
}
