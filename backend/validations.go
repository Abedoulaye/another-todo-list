package main

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}
	if len(email) > 254 {
		return errors.New("email must be less than 254 characters")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("Invalid email format")
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("password must be atleast 8 characters")
	}
	if len(password) > 72 {
		return errors.New("password cannot exceed 72 characters")
	}
	return nil
}

func validateNote(n Note) []string {
	var errs []string

	title := strings.TrimSpace(n.Title)
	if title == "" {
		errs = append(errs, "title is required")
	} else if len(title) > 255 {
		errs = append(errs, "name must be 255 characters or less")
	}

	if len(n.Body) > 5000 {
		errs = append(errs, "body must be 5000 characters or fewer")
	}

	return errs
}
