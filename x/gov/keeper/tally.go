package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	customtypes "github.com/initia-labs/initia/x/gov/types"
	stakingtypes "github.com/initia-labs/initia/x/mstaking/types"
)

// TODO: Break into several smaller functions for clarity

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx sdk.Context, proposal v1.Proposal) (quorumReached, passed bool, burnDeposits bool, tallyResults v1.TallyResult) {
	weights := keeper.sk.GetVotingPowerWeights(ctx)
	results := make(map[v1.VoteOption]sdk.Dec)
	results[v1.OptionYes] = math.LegacyZeroDec()
	results[v1.OptionAbstain] = math.LegacyZeroDec()
	results[v1.OptionNo] = math.LegacyZeroDec()
	results[v1.OptionNoWithVeto] = math.LegacyZeroDec()

	totalVotingPower := math.LegacyZeroDec()
	stakedVotingPower := math.ZeroInt()
	currValidators := make(map[string]customtypes.ValidatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		currValidators[validator.GetOperator().String()] = customtypes.NewValidatorGovInfo(
			validator.GetOperator(),
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			sdk.NewDecCoins(),
			v1.WeightedVoteOptions{},
		)

		votingPower, _ := stakingtypes.CalculateVotingPower(validator.GetBondedTokens(), weights)
		stakedVotingPower = stakedVotingPower.Add(votingPower)

		return false
	})

	keeper.IterateVotes(ctx, proposal.Id, func(vote v1.Vote) bool {
		// if validator, just record it in the map
		voter := sdk.MustAccAddressFromBech32(vote.Voter)

		valAddrStr := sdk.ValAddress(voter.Bytes()).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Options
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		keeper.sk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr().String()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will deduct voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares()...)
				currValidators[valAddrStr] = val

				// votingPower = delegation shares * bonded / total shares * denom weight
				votingPower := math.LegacyZeroDec()
				for _, share := range delegation.GetShares() {
					votingPower = votingPower.Add(
						share.Amount.
							MulInt(val.BondedTokens.AmountOf(share.Denom)).
							Quo(val.DelegatorShares.AmountOf(share.Denom)).
							Mul(weights.AmountOf(share.Denom)),
					)
				}

				for _, option := range vote.Options {
					subPower := votingPower.Mul(math.LegacyMustNewDecFromStr(option.Weight))
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})

		keeper.deleteVote(ctx, vote.ProposalId, voter)
		return false
	})

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := math.LegacyZeroDec()
		for _, share := range sharesAfterDeductions {
			votingPower = votingPower.Add(
				share.Amount.
					MulInt(val.BondedTokens.AmountOf(share.Denom)).
					Quo(val.DelegatorShares.AmountOf(share.Denom)).
					Mul(weights.AmountOf(share.Denom)),
			)
		}

		for _, option := range val.Vote {
			subPower := votingPower.Mul(math.LegacyMustNewDecFromStr(option.Weight))
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	params := keeper.GetParams(ctx)
	tallyResults = v1.NewTallyResultFromMap(results)

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	if stakedVotingPower.IsZero() {
		return false, false, false, tallyResults
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(math.LegacyNewDecFromInt(stakedVotingPower))
	if percentVoting.LT(math.LegacyMustNewDecFromStr(params.Quorum)) {
		return false, false, true, tallyResults
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return true, false, false, tallyResults
	}

	// If more than 1/3 of voters veto, proposal fails
	if results[v1.OptionNoWithVeto].Quo(totalVotingPower).GT(math.LegacyMustNewDecFromStr(params.VetoThreshold)) {
		return true, false, true, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[v1.OptionYes].Quo(totalVotingPower.Sub(results[v1.OptionAbstain])).GT(math.LegacyMustNewDecFromStr(params.Threshold)) {
		return true, true, false, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return true, false, false, tallyResults
}