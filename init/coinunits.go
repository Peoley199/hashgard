package init

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	gard  = "gard"  // 1 (base denom unit)
	mgard = "mgard" // 10^-3 (milli)
	ugard = "ugard" // 10^-6 (micro)
	ngard = "ngard" // 10^-9 (nano)
	pgard = "pgard"	// 10^-12 (pico)
	fgard = "fgard"	// 10^-15 (femto)
	agard = "agard"	// 10^-18 (atto)
)

func init() {
	InitNativeCoinUnits()
}

func InitNativeCoinUnits() {
	sdk.RegisterDenom(gard, sdk.OneDec())
	sdk.RegisterDenom(mgard, sdk.NewDecWithPrec(1, 3))
	sdk.RegisterDenom(ugard, sdk.NewDecWithPrec(1, 6))
	sdk.RegisterDenom(ngard, sdk.NewDecWithPrec(1, 9))
	sdk.RegisterDenom(pgard, sdk.NewDecWithPrec(1, 12))
	sdk.RegisterDenom(fgard, sdk.NewDecWithPrec(1, 15))
	sdk.RegisterDenom(agard, sdk.NewDecWithPrec(1, 18))
}