package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/hashgard/hashgard/x/box/params"

	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashgard/hashgard/x/box/client/queriers"
	boxqueriers "github.com/hashgard/hashgard/x/box/client/queriers"
	clientutils "github.com/hashgard/hashgard/x/box/client/utils"
	"github.com/hashgard/hashgard/x/box/errors"
	"github.com/hashgard/hashgard/x/box/msgs"
	"github.com/hashgard/hashgard/x/box/types"
	boxutils "github.com/hashgard/hashgard/x/box/utils"
	"github.com/spf13/cobra"
)

// GetCmdBoxDescription implements box a coin transaction command.
func GetCmdBoxDescription(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe [box-id] [description-file]",
		Args:    cobra.ExactArgs(2),
		Short:   "Describe a box",
		Long:    "Owner can add description of the box by owner, and the description need to be in json format. You can customize preferences or use recommended templates.",
		Example: "$ hashgardcli box describe boxab3jlxpt2ps path/description.json --from foo",
		RunE: func(cmd *cobra.Command, args []string) error {
			boxID := args[0]
			if err := boxutils.CheckBoxId(boxID); err != nil {
				return errors.Errorf(err)
			}
			txBldr, cliCtx, account, err := clientutils.GetCliContext(cdc)
			if err != nil {
				return err
			}
			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			buffer := bytes.Buffer{}
			err = json.Compact(&buffer, contents)
			if err != nil {
				return errors.ErrBoxDescriptionNotValid()
			}
			contents = buffer.Bytes()

			_, err = boxutils.BoxOwnerCheck(cdc, cliCtx, account, boxID)
			if err != nil {
				return err
			}
			msg := msgs.NewMsgBoxDescription(boxID, account.GetAddress(), contents)

			validateErr := msg.ValidateBasic()

			if validateErr != nil {
				return errors.Errorf(validateErr)
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}

	return cmd
}

// GetCmdBoxDisableFeature implements disable feature a box transaction command.
func GetCmdBoxDisableFeature(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable [box-id] [feature]",
		Args:  cobra.ExactArgs(2),
		Short: "Disable feature from a box",
		Long: fmt.Sprintf("Box Owner disabled the features:\n"+
			"%s:Box holder can trade the box", types.Trade),
		Example: fmt.Sprintf("$ hashgardcli box disable boxab3jlxpt2ps %s --from foo", types.Trade),
		RunE: func(cmd *cobra.Command, args []string) error {
			feature := args[1]

			_, ok := types.Features[feature]
			if !ok {
				return errors.Errorf(errors.ErrUnknownFeatures())
			}

			boxID := args[0]
			if err := boxutils.CheckBoxId(boxID); err != nil {
				return errors.Errorf(err)
			}
			txBldr, cliCtx, account, err := clientutils.GetCliContext(cdc)
			if err != nil {
				return err
			}
			boxInfo, err := boxutils.BoxOwnerCheck(cdc, cliCtx, account, boxID)
			if err != nil {
				return err
			}
			if feature == types.Trade && boxInfo.GetBoxType() == types.Lock {
				return errors.Errorf(errors.ErrNotSupportOperation())
			}

			msg := msgs.NewMsgBoxDisableFeature(boxID, account.GetAddress(), feature)
			validateErr := msg.ValidateBasic()
			if validateErr != nil {
				return errors.Errorf(validateErr)
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}

// GetCmdDepositToBox implements deposit to a box transaction command.
func GetCmdDepositToBox(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit-to [box-id] [amount]",
		Args:    cobra.ExactArgs(2),
		Short:   "Deposit to the box",
		Long:    "Deposit to the box",
		Example: "$ hashgardcli box deposit-to box174876e800 88888 --from foo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deposit(cdc, args, types.DepositTo)
		},
	}
	return cmd
}

// GetCmdFetchDepositFromBox implements fetch deposit from a box transaction command.
func GetCmdFetchDepositFromBox(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "deposit-fetch [box-id] [amount]",
		Args:    cobra.ExactArgs(2),
		Short:   "Fetch deposit from a deposit box",
		Long:    "Fetch deposit from a deposit box",
		Example: "$ hashgardcli box deposit-fetch box174876e800 88888 --from foo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deposit(cdc, args, types.Fetch)
		},
	}
	return cmd
}
func deposit(cdc *codec.Codec, args []string, operation string) error {

	boxID := args[0]
	if err := boxutils.CheckBoxId(boxID); err != nil {
		return errors.Errorf(err)
	}
	amountArg, ok := sdk.NewIntFromString(args[1])
	if !ok {
		return fmt.Errorf("Amount %s not a valid int, please input a valid amount", args[2])
	}
	txBldr, cliCtx, account, err := clientutils.GetCliContext(cdc)
	if err != nil {
		return err
	}
	boxInfo, err := boxutils.GetBoxByID(cdc, cliCtx, boxID)
	if err != nil {
		return err
	}
	if boxInfo.GetBoxStatus() != types.BoxDepositing {
		return errors.Errorf(errors.ErrNotAllowedOperation(boxInfo.GetBoxStatus()))
	}
	decimal, err := clientutils.GetCoinDecimal(cdc, cliCtx, boxInfo.GetTotalAmount().Token)
	if err != nil {
		return err
	}
	amount := boxutils.MulDecimals(boxutils.ParseCoin(boxInfo.GetTotalAmount().Token.Denom, amountArg), decimal)

	switch operation {
	case types.DepositTo:
		boxQueryParams := params.BoxQueryDepositListParams{
			BoxId: boxInfo.GetBoxId(),
		}
		// Query the box
		res, err := boxqueriers.QueryDepositList(boxQueryParams, cdc, cliCtx)
		if err != nil {
			return err
		}
		var boxs types.DepositBoxDepositInterestList
		cdc.MustUnmarshalJSON(res, &boxs)
		totalDeposit := sdk.ZeroInt()
		for i, box := range boxs {
			if box.Amount.IsZero() {
				continue
			}
			totalDeposit = totalDeposit.Add(boxs[i].Amount)
		}
		if amount.Add(totalDeposit).GT(boxInfo.GetTotalAmount().Token.Amount) {
			return errors.Errorf(errors.ErrNotEnoughAmount())
		}
		if err = checkAmountByDepositTo(amount, boxInfo); err != nil {
			return err
		}
	case types.Fetch:
		res, err := queriers.QueryDepositAmountFromDepositBox(boxID, account.GetAddress(), cliCtx)
		if err == nil {
			var boxDeposit types.BoxDeposit
			cdc.MustUnmarshalJSON(res, &boxDeposit)
			if boxDeposit.Amount.LT(amount) {
				return errors.Errorf(errors.ErrNotEnoughAmount())
			}
		}
	default:
		return errors.ErrNotSupportOperation()
	}
	msg := msgs.NewMsgBoxDeposit(boxID, account.GetAddress(), sdk.NewCoin(boxInfo.GetTotalAmount().Token.Denom, amount), operation)

	validateErr := msg.ValidateBasic()
	if validateErr != nil {
		return errors.Errorf(validateErr)
	}
	return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)

}

func checkAmountByDepositTo(amount sdk.Int, boxInfo types.Box) error {
	switch boxInfo.GetBoxType() {
	case types.Deposit:
		if !amount.Mod(boxInfo.GetDeposit().Price).IsZero() {
			return errors.ErrAmountNotValid(amount.String())
		}
		//if amount.Add(boxInfo.GetDeposit().TotalDeposit).GT(boxInfo.GetTotalAmount().Token.Amount) {
		//return errors.Errorf(errors.ErrNotEnoughAmount())
		//}
	case types.Future:
		total := sdk.ZeroInt()
		if boxInfo.GetFuture().Deposits != nil {
			for _, v := range boxInfo.GetFuture().Deposits {
				total = total.Add(v.Amount)
			}
		}
		if amount.Add(total).GT(boxInfo.GetTotalAmount().Token.Amount) {
			return errors.Errorf(errors.ErrNotEnoughAmount())
		}
	default:
		return errors.Errorf(errors.ErrNotSupportOperation())
	}
	return nil
}
