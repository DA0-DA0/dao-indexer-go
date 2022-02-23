package types

type CW3DAOGetConfigQuery struct {
	Data GetConfig `json:"get_config,omitempty"`
}

type ProposalCount struct{}

type CW3DAOGetProposalCountQuery struct {
	Data ProposalCount `json:"proposal_count,omitempty"`
}

type CW20TokenList struct{}
type CW3DAOGetCW20Tokens struct {
	Data CW20TokenList `json:"cw20_token_list,omitempty"`
}
