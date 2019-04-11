package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	issueCli "github.com/hashgard/hashgard/x/issue/client/cli"
	"github.com/hashgard/hashgard/x/issue/types"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

//New ModuleClient Instance
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetIssueCmd returns the issue commands for this module
func (mc ModuleClient) GetIssueCmd() *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Issue coin subcommands",
	}
	issueCmd.AddCommand(
		client.GetCommands(
			issueCli.GetCmdQueryIssue(mc.storeKey, mc.cdc),
			issueCli.GetCmdQueryIssues(mc.storeKey, mc.cdc),
		)...)
	issueCmd.AddCommand(client.LineBreak)
	issueCmd.AddCommand(
		client.PostCommands(
			issueCli.GetCmdIssueCreate(mc.cdc),
			issueCli.GetCmdIssueMint(mc.cdc),
			issueCli.GetCmdIssueBurn(mc.cdc),
			issueCli.GetCmdIssueFinishMinting(mc.cdc),
		)...)

	return issueCmd
}
