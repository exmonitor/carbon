package email

import "fmt"

func Send(email string, message string) error {

	fmt.Printf("<< Fake email notification sent to %s\n", email)
	return nil
}
