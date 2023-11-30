package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/initia/x/mstaking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinPos  = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinZero = sdk.NewCoins()
)

func TestMsgDecode(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// firstly we start testing the pubkey serialization

	pk1bz, err := cdc.MarshalInterface(pk1)
	require.NoError(t, err)
	var pkUnmarshaled cryptotypes.PubKey
	err = cdc.UnmarshalInterface(pk1bz, &pkUnmarshaled)
	require.NoError(t, err)
	require.True(t, pk1.Equals(pkUnmarshaled.(*ed25519.PubKey)))

	// now let's try to serialize the whole message

	commission1 := types.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
	msg, err := types.NewMsgCreateValidator(valAddr1, pk1, coinPos, types.Description{}, commission1)
	require.NoError(t, err)
	msgSerialized, err := cdc.MarshalInterface(msg)
	require.NoError(t, err)

	var msgUnmarshaled sdk.Msg
	err = cdc.UnmarshalInterface(msgSerialized, &msgUnmarshaled)
	require.NoError(t, err)
	msg2, ok := msgUnmarshaled.(*types.MsgCreateValidator)
	require.True(t, ok)
	require.True(t, msg.Amount.IsEqual(msg2.Amount))
	require.True(t, msg.Pubkey.Equal(msg2.Pubkey))
}

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidator(t *testing.T) {
	commission1 := types.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
	commission2 := types.NewCommissionRates(sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5))

	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		CommissionRates                                            types.CommissionRates
		validatorAddr                                              sdk.ValAddress
		pubkey                                                     cryptotypes.PubKey
		bond                                                       sdk.Coins
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", commission1, valAddr1, pk1, coinPos, true},
		{"partial description", "", "", "c", "", "", commission1, valAddr1, pk1, coinPos, true},
		{"empty description", "", "", "", "", "", commission2, valAddr1, pk1, coinPos, false},
		{"empty address", "a", "b", "c", "d", "e", commission2, emptyAddr, pk1, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", "e", commission1, valAddr1, emptyPubkey, coinPos, false},
		{"empty bond", "a", "b", "c", "d", "e", commission2, valAddr1, pk1, coinZero, false},
		{"nil bond", "a", "b", "c", "d", "e", commission2, valAddr1, pk1, sdk.Coins{}, false},
	}

	for _, tc := range tests {
		description := types.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		msg, err := types.NewMsgCreateValidator(tc.validatorAddr, tc.pubkey, tc.bond, description, tc.CommissionRates)
		require.NoError(t, err)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgEditValidator
func TestMsgEditValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		validatorAddr                                              sdk.ValAddress
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", valAddr1, true},
		{"partial description", "", "", "c", "", "", valAddr1, true},
		{"empty description", "", "", "", "", "", valAddr1, false},
		{"empty address", "a", "b", "c", "d", "e", emptyAddr, false},
	}

	for _, tc := range tests {
		description := types.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		newRate := math.LegacyZeroDec()

		msg := types.NewMsgEditValidator(tc.validatorAddr, description, &newRate)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgDelegate
func TestMsgDelegate(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		bond          sdk.Coins
		expectPass    bool
	}{
		{"basic good", sdk.AccAddress(valAddr1), valAddr2, coinPos, true},
		{"self bond", sdk.AccAddress(valAddr1), valAddr1, coinPos, true},
		{"empty delegator", sdk.AccAddress(emptyAddr), valAddr1, coinPos, false},
		{"empty validator", sdk.AccAddress(valAddr1), emptyAddr, coinPos, false},
		{"empty bond", sdk.AccAddress(valAddr1), valAddr2, coinZero, false},
		{"nil bold", sdk.AccAddress(valAddr1), valAddr2, sdk.Coins{}, false},
	}

	for _, tc := range tests {
		msg := types.NewMsgDelegate(tc.delegatorAddr, tc.validatorAddr, tc.bond)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgBeginRedelegate(t *testing.T) {
	tests := []struct {
		name             string
		delegatorAddr    sdk.AccAddress
		validatorSrcAddr sdk.ValAddress
		validatorDstAddr sdk.ValAddress
		amount           sdk.Coins
		expectPass       bool
	}{
		{"regular", sdk.AccAddress(valAddr1), valAddr2, valAddr3, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), true},
		{"zero amount", sdk.AccAddress(valAddr1), valAddr2, valAddr3, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)), false},
		{"nil amount", sdk.AccAddress(valAddr1), valAddr2, valAddr3, sdk.Coins{}, false},
		{"empty delegator", sdk.AccAddress(emptyAddr), valAddr1, valAddr3, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), false},
		{"empty source validator", sdk.AccAddress(valAddr1), emptyAddr, valAddr3, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), false},
		{"empty destination validator", sdk.AccAddress(valAddr1), valAddr2, emptyAddr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), false},
	}

	for _, tc := range tests {
		msg := types.NewMsgBeginRedelegate(tc.delegatorAddr, tc.validatorSrcAddr, tc.validatorDstAddr, tc.amount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgUndelegate(t *testing.T) {
	tests := []struct {
		name          string
		delegatorAddr sdk.AccAddress
		validatorAddr sdk.ValAddress
		amount        sdk.Coins
		expectPass    bool
	}{
		{"regular", sdk.AccAddress(valAddr1), valAddr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), true},
		{"zero amount", sdk.AccAddress(valAddr1), valAddr2, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)), false},
		{"nil amount", sdk.AccAddress(valAddr1), valAddr2, sdk.Coins{}, false},
		{"empty delegator", sdk.AccAddress(emptyAddr), valAddr1, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), false},
		{"empty validator", sdk.AccAddress(valAddr1), emptyAddr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)), false},
	}

	for _, tc := range tests {
		msg := types.NewMsgUndelegate(tc.delegatorAddr, tc.validatorAddr, tc.amount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}