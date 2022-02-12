package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ExecMsg struct {
	Sender  string
	Address string
	Json    string
	Funds   sdk.Coins
}

// NewContract instance
func ExecutedMessage(sender string, address string, funds sdk.Coins, json string) ExecMsg {
	return ExecMsg{
		Sender:  sender,
		Address: address,
		Funds:   funds,
		Json:    json,
	}
}
