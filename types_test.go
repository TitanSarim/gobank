package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T){
	acc, err := NewAccount("a@xyc.com", "a", "b", "sarim")

    assert.Nil(t, err)

	fmt.Printf("%+v\n", acc)
}