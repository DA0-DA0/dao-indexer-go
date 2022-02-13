package wasm

import (
	"context"
	"fmt"

	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/disperze/wasmx/types"
	juno "github.com/forbole/juno/v2/types"
)

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *wasmtypes.MsgStoreCode:
		return m.handleMsgStoreCode(tx, index, cosmosMsg)
	case *wasmtypes.MsgInstantiateContract:
		return m.handleMsgInstantiateContract(tx, index, cosmosMsg)
	case *wasmtypes.MsgMigrateContract:
		return m.handleMsgMigrateContract(tx, index, cosmosMsg)
	case *wasmtypes.MsgClearAdmin:
		return m.handleMsgClearAdmin(tx, index, cosmosMsg)
	case *wasmtypes.MsgUpdateAdmin:
		return m.handleMsgUpdateAdmin(tx, index, cosmosMsg)
	case *wasmtypes.MsgExecuteContract:
		return m.handleMsgExecuteContract(tx, index, cosmosMsg)
	}

	return nil
}

func (m *Module) handleMsgExecuteContract(tx *juno.Tx, index int, msg *wasmtypes.MsgExecuteContract) error {
	json := string(msg.Msg)
	address := msg.Contract
	sender := msg.Sender
	funds := msg.Funds
	execMsg := types.ExecutedMessage(sender, address, funds, json)
	return m.db.SaveExec(execMsg)
}

func (m *Module) handleMsgStoreCode(tx *juno.Tx, index int, msg *wasmtypes.MsgStoreCode) error {
	event, err := tx.FindEventByType(index, wasmtypes.EventTypeStoreCode)
	if err != nil {
		return err
	}

	codeID, err := tx.FindAttributeByKey(event, wasmtypes.AttributeKeyCodeID)
	if err != nil {
		return err
	}

	code := types.NewCode(codeID, msg.Sender, tx.Timestamp, tx.Height)

	return m.db.SaveCode(code)
}

func (m *Module) handleMsgInstantiateContract(tx *juno.Tx, index int, msg *wasmtypes.MsgInstantiateContract) error {
	contracts, err := GetAllContracts(tx, index, wasmtypes.EventTypeInstantiate)
	if err != nil {
		return err
	}

	if len(contracts) == 0 {
		return fmt.Errorf("no contract address found")
	}

	createdAt := &wasmtypes.AbsoluteTxPosition{
		BlockHeight: uint64(tx.Height),
		TxIndex:     uint64(index),
	}
	ctx := context.Background()
	for _, contractAddress := range contracts {
		response, err := m.client.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
			Address: contractAddress,
		})
		if err != nil {
			return err
		}

		creator, _ := sdk.AccAddressFromBech32(response.Creator)
		var admin sdk.AccAddress
		if response.Admin != "" {
			admin, _ = sdk.AccAddressFromBech32(response.Admin)
		}

		json_string := string(msg.Msg)

		type Duration struct {
			Time int
		}

		type Balance struct {
			Address             string
			Amount              int64
			StakeContractCodeId int
			InitialDaoBalance   int64
			UnstakingDuration   Duration
		}

		type Cw20Msg struct {
			Name     string
			Symbol   string
			Decimals int
			Balances []Balance
		}
		type InstantiateCw20 struct {
			Cw20CodeId          int32 `json:"cw20_code_id"`
			Label               string
			Msg                 Cw20Msg `json:"msg"`
			StakeContractCodeId string
		}

		type GovToken struct {
			InstantiateCw20 InstantiateCw20 `json:"instantiate_new_cw20"`
		}

		type DaoInstantiateContractMessage struct {
			Name        string
			Description string
			GovToken    GovToken `json:"gov_token"`
		}
		/*
			"{\"name\":\"d2\",\"description\":\"d2d\",
			\"gov_token\":{\"instantiate_new_cw20\":
			{\"cw20_code_id\":1,\"label\":\"db_t\",\"msg\":
			{\"name\":\"db_t\",\"symbol\":\"dbt\",\"decimals\":6,\"
			initial_balances\":[{
				\"address\":\"juno1mudcxmlg5gxqkwuywuedql79wgy3m02rtqac8a\",
				\"amount\":\"1000000\"}]},
				\"stake_contract_code_id\":5,
				\"initial_dao_balance\":\"5000000\",\
				"unstaking_duration\":{
					\"time\":0
					}
					}},
				\"threshold\":{
					\"absolute_percentage\":{
						\"percentage\":\"0.75\"}},
						\"max_voting_period\":{\"time\":604800},
						\"proposal_deposit_amount\":\"0\",\"refund_failed_proposals\":true}"
		*/
		var contractMessage DaoInstantiateContractMessage
		unmarshallErr := json.Unmarshal(msg.Msg, &contractMessage)
		fmt.Println(unmarshallErr)
		govToken := contractMessage.GovToken
		fmt.Println(govToken)

		contractInfo := wasmtypes.NewContractInfo(response.CodeID, creator, admin, response.Label, createdAt)
		contract := types.NewContract(&contractInfo, contractAddress, tx.Timestamp, json_string)

		if err = m.db.SaveContract(contract); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) handleMsgMigrateContract(tx *juno.Tx, index int, msg *wasmtypes.MsgMigrateContract) error {

	return m.db.SaveContractCodeID(msg.Contract, msg.CodeID)
}

func (m *Module) handleMsgClearAdmin(tx *juno.Tx, index int, msg *wasmtypes.MsgClearAdmin) error {

	return m.db.UpdateContractAdmin(msg.Contract, "")
}

func (m *Module) handleMsgUpdateAdmin(tx *juno.Tx, index int, msg *wasmtypes.MsgUpdateAdmin) error {

	return m.db.UpdateContractAdmin(msg.Contract, msg.NewAdmin)
}

func GetAllContracts(tx *juno.Tx, index int, eventType string) ([]string, error) {
	contracts := []string{}
	event, err := tx.FindEventByType(index, eventType)
	if err != nil {
		return contracts, err
	}

	for _, attr := range event.Attributes {
		if attr.Key == wasmtypes.AttributeKeyContractAddr {
			contracts = append(contracts, attr.Value)
		}
	}

	return contracts, nil
}
