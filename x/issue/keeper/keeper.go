package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/hashgard/hashgard/x/issue/errors"
	issueparams "github.com/hashgard/hashgard/x/issue/params"
	"github.com/hashgard/hashgard/x/issue/types"
	"github.com/hashgard/hashgard/x/issue/utils"
)

// Parameter store key
var (
	ParamStoreKeyIssueParams = []byte("issueparams")
)

// Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyIssueParams, issueparams.IssueConfigParams{},
	)
}

// Issue Keeper
type Keeper struct {
	// The reference to the Param Keeper to get and set Global Params
	paramsKeeper params.Keeper
	// The reference to the Paramstore to get and set issue specific params
	paramSpace params.Subspace
	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey
	// The reference to the CoinKeeper to modify balances
	ck BankKeeper
	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec
	// Reserved codespace
	codespace sdk.CodespaceType
}

//Get issue codec
func (keeper Keeper) Getcdc() *codec.Codec {
	return keeper.cdc
}

//New issue keeper Instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramsKeeper params.Keeper,
	paramSpace params.Subspace, ck BankKeeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:     key,
		paramsKeeper: paramsKeeper,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		ck:           ck,
		cdc:          cdc,
		codespace:    codespace,
	}
}

//Returns issue by issueID
func (keeper Keeper) GetIssue(ctx sdk.Context, issueID string) *types.CoinIssueInfo {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyIssuer(issueID))
	if len(bz) == 0 {
		return nil
	}
	var coinIssueInfo types.CoinIssueInfo
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &coinIssueInfo)
	return &coinIssueInfo
}

//Returns issues by accAddress
func (keeper Keeper) GetIssues(ctx sdk.Context, accAddress string) []*types.CoinIssueInfo {

	issueIDs := keeper.GetAddressIssues(ctx, accAddress)
	length := len(issueIDs)
	if length == 0 {
		return []*types.CoinIssueInfo{}
	}
	issues := make([]*types.CoinIssueInfo, 0, length)

	for _, v := range issueIDs {
		issues = append(issues, keeper.GetIssue(ctx, v))
	}

	return issues
}
func (keeper Keeper) SearchIssues(ctx sdk.Context, symbol string) []*types.CoinIssueInfo {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, KeySymbolIssues(symbol))
	defer iterator.Close()
	list := make([]*types.CoinIssueInfo, 0, 1)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		if len(bz) == 0 {
			continue
		}
		issueIDs := make([]string, 0, 1)
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issueIDs)

		for _, v := range issueIDs {
			list = append(list, keeper.GetIssue(ctx, v))
		}
	}
	return list
}
func (keeper Keeper) List(ctx sdk.Context, params issueparams.IssueQueryParams) []*types.CoinIssueInfo {
	store := ctx.KVStore(keeper.storeKey)
	startIssueId := params.StartIssueId
	endIssueId := startIssueId

	if len(startIssueId) == 0 {
		endIssueId = KeyIssueIdStr(types.CoinIssueMaxId)
		startIssueId = KeyIssueIdStr(types.CoinIssueMinId - 1)
	} else {
		startIssueId = KeyIssueIdStr(types.CoinIssueMinId - 1)
	}

	iterator := store.ReverseIterator(KeyIssuer(startIssueId), KeyIssuer(endIssueId))
	defer iterator.Close()
	list := make([]*types.CoinIssueInfo, 0, params.Limit)
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		if len(bz) == 0 {
			continue
		}
		var info types.CoinIssueInfo
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &info)
		list = append(list, &info)
		if len(list) >= params.Limit {
			break
		}
	}
	return list
}

//Add a issue
func (keeper Keeper) AddIssue(ctx sdk.Context, coinIssueInfo *types.CoinIssueInfo) (sdk.Coins, sdk.Tags, sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	id, err := keeper.getNewIssueID(store)
	if err != nil {
		return nil, nil, err
	}
	issueID := KeyIssueIdStr(id)

	//issueID := keeper.getIssueId(store)
	coinIssueInfo.IssueId = issueID
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo)

	store.Set(KeyIssuer(coinIssueInfo.IssueId), bz)

	issueIDs := keeper.GetAddressIssues(ctx, coinIssueInfo.GetIssuer().String())
	issueIDs = append(issueIDs, coinIssueInfo.IssueId)
	keeper.setAddressIssues(ctx, coinIssueInfo.GetIssuer().String(), issueIDs)

	issueIDs = keeper.GetSymbolIssues(ctx, coinIssueInfo.Symbol)
	issueIDs = append(issueIDs, coinIssueInfo.IssueId)
	keeper.setSymbolIssues(ctx, coinIssueInfo.Symbol, issueIDs)

	coin := sdk.Coin{Denom: coinIssueInfo.IssueId, Amount: coinIssueInfo.TotalSupply}
	coins, tags, err := keeper.ck.AddCoins(ctx, coinIssueInfo.Owner, sdk.NewCoins(coin))

	return coins, tags, err
}
func (keeper Keeper) getIssueByOwner(ctx sdk.Context, operator sdk.AccAddress, issueID string) (*types.CoinIssueInfo, sdk.Error) {
	coinIssueInfo := keeper.GetIssue(ctx, issueID)
	if coinIssueInfo == nil {
		return nil, errors.ErrUnknownIssue(issueID)
	}
	if !coinIssueInfo.Owner.Equals(operator) {
		return nil, errors.ErrOwnerMismatch(issueID)
	}
	return coinIssueInfo, nil
}

//Finished Minting a coin
func (keeper Keeper) FinishMinting(ctx sdk.Context, operator sdk.AccAddress, issueID string) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)
	if err != nil {
		return err
	}

	if coinIssueInfo.MintingFinished {
		return nil
	}
	coinIssueInfo.MintingFinished = true
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(issueID), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//BurnOff a coin
func (keeper Keeper) BurnOff(ctx sdk.Context, operator sdk.AccAddress, issueID string) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)
	if err != nil {
		return err
	}

	if coinIssueInfo.BurnOff {
		return nil
	}
	coinIssueInfo.BurnOff = true
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(issueID), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//BurnFromOff a coin
func (keeper Keeper) BurnFromOff(ctx sdk.Context, operator sdk.AccAddress, issueID string) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)
	if err != nil {
		return err
	}

	if coinIssueInfo.BurnFromOff {
		return nil
	}
	coinIssueInfo.BurnFromOff = true
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(issueID), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//BurnAnyOff a coin
func (keeper Keeper) BurnAnyOff(ctx sdk.Context, operator sdk.AccAddress, issueID string) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)
	if err != nil {
		return err
	}

	if coinIssueInfo.BurnAnyOff {
		return nil
	}
	coinIssueInfo.BurnAnyOff = true
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(issueID), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//Can mint a coin
func (keeper Keeper) CanMint(ctx sdk.Context, issueID string) bool {
	coinIssueInfo := keeper.GetIssue(ctx, issueID)
	return !coinIssueInfo.MintingFinished
}

//Mint a coin
func (keeper Keeper) Mint(ctx sdk.Context, issueID string, amount sdk.Int, operator sdk.AccAddress, to sdk.AccAddress) (sdk.Coins, sdk.Tags, sdk.Error) {

	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)

	if err != nil {
		return nil, nil, err
	}

	if coinIssueInfo.MintingFinished {
		return nil, nil, errors.ErrCanNotMint(issueID)
	}
	if utils.QuoDecimals(coinIssueInfo.TotalSupply.Add(amount), coinIssueInfo.Decimals).GT(types.CoinMaxTotalSupply) {
		return nil, nil, errors.ErrCoinTotalSupplyMaxValueNotValid()
	}

	coin := sdk.Coin{Denom: coinIssueInfo.IssueId, Amount: amount}
	coins, tags, err := keeper.ck.AddCoins(ctx, to, sdk.NewCoins(coin))
	if err != nil {
		return coins, tags, err
	}
	coinIssueInfo.TotalSupply = coinIssueInfo.TotalSupply.Add(amount)
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(coinIssueInfo.IssueId), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))

	return coins, tags, err
}

//Burn a coin
func (keeper Keeper) Burn(ctx sdk.Context, issueID string, amount sdk.Int, operator sdk.AccAddress) (sdk.Coins, sdk.Tags, sdk.Error) {
	coinIssueInfo := keeper.GetIssue(ctx, issueID)

	if coinIssueInfo == nil {
		return nil, nil, errors.ErrUnknownIssue(issueID)
	}
	if coinIssueInfo.GetBurnOff() {
		return nil, nil, errors.ErrCanNotBurn(issueID)
	}
	if !coinIssueInfo.Owner.Equals(operator) {
		return nil, nil, errors.ErrOwnerMismatch(issueID)
	}

	return keeper.burn(ctx, coinIssueInfo, amount, operator)
}
func (keeper Keeper) burn(ctx sdk.Context, coinIssueInfo *types.CoinIssueInfo, amount sdk.Int, who sdk.AccAddress) (sdk.Coins, sdk.Tags, sdk.Error) {
	coin := sdk.Coin{Denom: coinIssueInfo.IssueId, Amount: amount}
	coins, tags, err := keeper.ck.SubtractCoins(ctx, who, sdk.NewCoins(coin))
	if err != nil {
		return nil, nil, err
	}

	coinIssueInfo.TotalSupply = coinIssueInfo.TotalSupply.Sub(amount)
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(coinIssueInfo.IssueId), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return coins, tags, nil
}

//Burn a coin from address
func (keeper Keeper) BurnFrom(ctx sdk.Context, issueID string, amount sdk.Int, operator sdk.AccAddress, burnfrom sdk.AccAddress) (sdk.Coins, sdk.Tags, sdk.Error) {
	coinIssueInfo := keeper.GetIssue(ctx, issueID)

	if coinIssueInfo == nil {
		return nil, nil, errors.ErrUnknownIssue(issueID)
	}

	if operator.Equals(coinIssueInfo.GetOwner()) {
		if coinIssueInfo.GetBurnAnyOff() {
			return nil, nil, errors.ErrCanNotBurn(issueID)
		}
	} else {
		if coinIssueInfo.GetBurnFromOff() {
			return nil, nil, errors.ErrCanNotBurn(issueID)
		}
		if !burnfrom.Equals(operator) {
			return nil, nil, errors.ErrCanNotBurn(issueID)
		}
	}

	return keeper.burn(ctx, coinIssueInfo, amount, burnfrom)
}
func (keeper Keeper) SetIssueDescription(ctx sdk.Context, issueID string, operator sdk.AccAddress, description []byte) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)

	if err != nil {
		return err
	}

	coinIssueInfo.Description = string(description)
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(coinIssueInfo.IssueId), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//TransferOwnership
func (keeper Keeper) TransferOwnership(ctx sdk.Context, issueID string, operator sdk.AccAddress, to sdk.AccAddress) sdk.Error {
	coinIssueInfo, err := keeper.getIssueByOwner(ctx, operator, issueID)

	if err != nil {
		return err
	}

	coinIssueInfo.Owner = to
	store := ctx.KVStore(keeper.storeKey)
	store.Set(KeyIssuer(coinIssueInfo.IssueId), keeper.cdc.MustMarshalBinaryLengthPrefixed(coinIssueInfo))
	return nil
}

//Send coins
func (keeper Keeper) SendCoins(ctx sdk.Context,
	fromAddr sdk.AccAddress, toAddr sdk.AccAddress,
	amt sdk.Coins) (sdk.Tags, sdk.Error) {
	return keeper.ck.SendCoins(ctx, fromAddr, toAddr, amt)
}

//Get address from a issue
func (keeper Keeper) GetAddressIssues(ctx sdk.Context, accAddress string) (issueIDs []string) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyAddressIssues(accAddress))
	if bz == nil {
		return []string{}
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issueIDs)
	return issueIDs
}

//Set issueIDs to a issue
func (keeper Keeper) setAddressIssues(ctx sdk.Context, accAddress string, issueIDs []string) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(issueIDs)
	store.Set(KeyAddressIssues(accAddress), bz)
}

//Get issueIDs from a issue
func (keeper Keeper) GetSymbolIssues(ctx sdk.Context, symbol string) (issueIDs []string) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeySymbolIssues(symbol))
	if bz == nil {
		return []string{}
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issueIDs)
	return issueIDs
}

//Set symbol to a issue
func (keeper Keeper) setSymbolIssues(ctx sdk.Context, symbol string, issueIDs []string) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(issueIDs)
	store.Set(KeySymbolIssues(symbol), bz)
}

// Params
// Returns the current issueConfigParams from the global param store
func (keeper Keeper) GetIssueConfigParams(ctx sdk.Context) issueparams.IssueConfigParams {
	var issueConfigParams issueparams.IssueConfigParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyIssueParams, &issueConfigParams)
	return issueConfigParams
}

//Set issueConfigParams
func (keeper Keeper) SetIssueConfigParams(ctx sdk.Context, issueConfigParams issueparams.IssueConfigParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyIssueParams, &issueConfigParams)
}

// Set the initial issueCount
func (keeper Keeper) SetInitialIssueStartingIssueId(ctx sdk.Context, issueID uint64) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextIssueID)
	if bz != nil {
		return sdk.NewError(keeper.codespace, types.CodeInvalidGenesis, "Initial IssueId already set")
	}
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(issueID)
	store.Set(KeyNextIssueID, bz)
	return nil
}

// Get the last used issueID
func (keeper Keeper) GetLastIssueID(ctx sdk.Context) (issueID uint64) {
	issueID, err := keeper.PeekCurrentIssueID(ctx)
	if err != nil {
		return 0
	}
	issueID--
	return
}

// Gets the next available issueID and increments it
func (keeper Keeper) getNewIssueID(store sdk.KVStore) (issueID uint64, err sdk.Error) {
	bz := store.Get(KeyNextIssueID)
	if bz == nil {
		return 0, sdk.NewError(keeper.codespace, types.CodeInvalidGenesis, "InitialIssueID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issueID)
	bz = keeper.cdc.MustMarshalBinaryLengthPrefixed(issueID + 1)
	store.Set(KeyNextIssueID, bz)
	return issueID, nil
}

// Peeks the next available IssueID without incrementing it
func (keeper Keeper) PeekCurrentIssueID(ctx sdk.Context) (issueID uint64, err sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(KeyNextIssueID)
	if bz == nil {
		return 0, sdk.NewError(keeper.codespace, types.CodeInvalidGenesis, "InitialIssueID never set")
	}
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issueID)
	return issueID, nil
}
