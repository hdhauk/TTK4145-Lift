package main

import uuid "github.com/satori/go.uuid"

func makeUUID() string {
	return uuid.NewV4().String()
}
