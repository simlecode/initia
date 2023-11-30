package app

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/math"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/streaming"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	solomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	ibcnfttransfer "github.com/initia-labs/initia/x/ibc/nft-transfer"
	ibcnfttransferkeeper "github.com/initia-labs/initia/x/ibc/nft-transfer/keeper"
	ibcnfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
	ibcperm "github.com/initia-labs/initia/x/ibc/perm"
	ibcpermkeeper "github.com/initia-labs/initia/x/ibc/perm/keeper"
	ibcpermtypes "github.com/initia-labs/initia/x/ibc/perm/types"
	ibctestingtypes "github.com/initia-labs/initia/x/ibc/testing/types"
	icaauth "github.com/initia-labs/initia/x/intertx"
	icaauthkeeper "github.com/initia-labs/initia/x/intertx/keeper"
	icaauthtypes "github.com/initia-labs/initia/x/intertx/types"
	"github.com/strangelove-ventures/packet-forward-middleware/v7/router"
	routerkeeper "github.com/strangelove-ventures/packet-forward-middleware/v7/router/keeper"
	routertypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"

	// this line is used by starport scaffolding # stargate/app/moduleImport

	appante "github.com/initia-labs/initia/app/ante"
	apphook "github.com/initia-labs/initia/app/hook"
	applanes "github.com/initia-labs/initia/app/lanes"
	authzmodule "github.com/initia-labs/initia/x/authz/module"
	"github.com/initia-labs/initia/x/bank"
	bankkeeper "github.com/initia-labs/initia/x/bank/keeper"
	distr "github.com/initia-labs/initia/x/distribution"
	distrkeeper "github.com/initia-labs/initia/x/distribution/keeper"
	"github.com/initia-labs/initia/x/evidence"
	evidencekeeper "github.com/initia-labs/initia/x/evidence/keeper"
	"github.com/initia-labs/initia/x/genutil"
	"github.com/initia-labs/initia/x/gov"
	govkeeper "github.com/initia-labs/initia/x/gov/keeper"
	"github.com/initia-labs/initia/x/move"
	moveconfig "github.com/initia-labs/initia/x/move/config"
	moveibcmiddleware "github.com/initia-labs/initia/x/move/ibc-middleware"
	movekeeper "github.com/initia-labs/initia/x/move/keeper"
	movetypes "github.com/initia-labs/initia/x/move/types"
	staking "github.com/initia-labs/initia/x/mstaking"
	stakingkeeper "github.com/initia-labs/initia/x/mstaking/keeper"
	stakingtypes "github.com/initia-labs/initia/x/mstaking/types"
	reward "github.com/initia-labs/initia/x/reward"
	rewardkeeper "github.com/initia-labs/initia/x/reward/keeper"
	rewardtypes "github.com/initia-labs/initia/x/reward/types"
	"github.com/initia-labs/initia/x/slashing"
	slashingkeeper "github.com/initia-labs/initia/x/slashing/keeper"

	blockabci "github.com/skip-mev/block-sdk/abci"
	signer_extraction "github.com/skip-mev/block-sdk/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/block"
	blockbase "github.com/skip-mev/block-sdk/block/base"
	baselane "github.com/skip-mev/block-sdk/lanes/base"
	freelane "github.com/skip-mev/block-sdk/lanes/free"
	"github.com/skip-mev/block-sdk/lanes/mev"
	"github.com/skip-mev/block-sdk/x/auction"
	auctionante "github.com/skip-mev/block-sdk/x/auction/ante"
	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/initia-labs/OPinit/x/ophost"
	ophostkeeper "github.com/initia-labs/OPinit/x/ophost/keeper"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/initia-labs/initia/client/docs/statik"
)

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		reward.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()...),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		consensus.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		solomachine.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ibcnfttransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		icaauth.AppModuleBasic{},
		ibcfee.AppModuleBasic{},
		router.AppModuleBasic{},
		ibcperm.AppModuleBasic{},
		move.AppModuleBasic{},
		auction.AppModuleBasic{},
		ophost.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:      nil,
		distrtypes.ModuleName:           nil,
		icatypes.ModuleName:             nil,
		ibcfeetypes.ModuleName:          nil,
		rewardtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:             {authtypes.Burner},
		ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
		movetypes.MoveStakingModuleName: nil,
		// x/auction's module account must be instantiated upon genesis to accrue auction rewards not
		// distributed to proposers
		auctiontypes.ModuleName: nil,

		// this is only for testing
		authtypes.Minter: {authtypes.Minter},
	}
)

var (
	_ servertypes.Application = (*InitiaApp)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+AppName)
}

// InitiaApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type InitiaApp struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         *authkeeper.AccountKeeper
	BankKeeper            *bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        *slashingkeeper.Keeper
	RewardKeeper          *rewardkeeper.Keeper
	DistrKeeper           *distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          *paramskeeper.Keeper
	GroupKeeper           *groupkeeper.Keeper
	ConsensusParamsKeeper *consensusparamkeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        *evidencekeeper.Keeper
	TransferKeeper        *ibctransferkeeper.Keeper
	NftTransferKeeper     *ibcnfttransferkeeper.Keeper
	AuthzKeeper           *authzkeeper.Keeper
	FeeGrantKeeper        *feegrantkeeper.Keeper
	ICAHostKeeper         *icahostkeeper.Keeper
	ICAControllerKeeper   *icacontrollerkeeper.Keeper
	ICAAuthKeeper         *icaauthkeeper.Keeper
	IBCFeeKeeper          *ibcfeekeeper.Keeper
	RouterKeeper          *routerkeeper.Keeper // Router Keeper must be a pointer in the app, so we can SetTransferKeeper on it correctly
	IBCPermKeeper         *ibcpermkeeper.Keeper
	MoveKeeper            *movekeeper.Keeper
	AuctionKeeper         *auctionkeeper.Keeper // x/auction keeper used to process bids for TOB auctions
	OPHostKeeper          *ophostkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedNftTransferKeeper   capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedICAAuthKeeper       capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager

	// the configurator
	configurator module.Configurator

	// Override of BaseApp's CheckTx
	checkTxHandler mev.CheckTx
}

// NewInitiaApp returns a reference to an initialized Initia.
func NewInitiaApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	moveConfig moveconfig.MoveConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *InitiaApp {
	encodingConfig := MakeEncodingConfig()

	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, crisistypes.StoreKey,
		rewardtypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, group.StoreKey, paramstypes.StoreKey, consensusparamtypes.StoreKey,
		ibcexported.StoreKey, upgradetypes.StoreKey, evidencetypes.StoreKey,
		ibctransfertypes.StoreKey, ibcnfttransfertypes.StoreKey, capabilitytypes.StoreKey,
		authzkeeper.StoreKey, feegrant.StoreKey, icahosttypes.StoreKey,
		icacontrollertypes.StoreKey, icaauthtypes.StoreKey, ibcfeetypes.StoreKey,
		routertypes.StoreKey, ibcpermtypes.StoreKey, movetypes.StoreKey, auctiontypes.StoreKey,
		ophosttypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// load state streaming if enabled
	if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, logger, keys); err != nil {
		logger.Error("failed to load state streaming", "err", err)
		os.Exit(1)
	}

	app := &InitiaApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	consensusParamKeeper := consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	app.ConsensusParamsKeeper = &consensusParamKeeper
	bApp.SetParamStore(app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedNftTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibcnfttransfertypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAAuthKeeper := app.CapabilityKeeper.ScopeToModule(icaauthtypes.ModuleName)

	app.CapabilityKeeper.Seal()

	// add keepers
	moveKeeper := &movekeeper.Keeper{}

	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.AccountKeeper = &accountKeeper

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		movekeeper.NewMoveBankKeeper(moveKeeper),
		app.ModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.BankKeeper = &bankKeeper

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		movekeeper.NewVotingPowerKeeper(moveKeeper),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.StakingKeeper = &stakingKeeper

	rewardKeeper := rewardkeeper.NewKeeper(
		appCodec, keys[rewardtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.RewardKeeper = &rewardKeeper

	distrKeeper := distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		movekeeper.NewDexKeeper(moveKeeper),
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.DistrKeeper = &distrKeeper

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.SlashingKeeper = &slashingKeeper

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	i := 0
	moduleAddrs := make([]sdk.AccAddress, len(maccPerms))
	for name := range maccPerms {
		moduleAddrs[i] = authtypes.NewModuleAddress(name)
		i += 1
	}

	feeGrantKeeper := feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.FeeGrantKeeper = &feeGrantKeeper

	authzKeeper := authzkeeper.NewKeeper(keys[authzkeeper.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper)
	app.AuthzKeeper = &authzKeeper

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	app.EvidenceKeeper = evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)

	groupConfig := group.DefaultConfig()
	groupKeeper := groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)
	app.GroupKeeper = &groupKeeper

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)

	ibcFeeKeeper := ibcfeekeeper.NewKeeper(
		appCodec,
		keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)
	app.IBCFeeKeeper = &ibcFeeKeeper

	app.IBCPermKeeper = ibcpermkeeper.NewKeeper(
		appCodec,
		keys[ibcpermtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	////////////////////////////
	// Transfer configuration //
	////////////////////////////
	// Send   : transfer -> router -> fee    -> channel
	// Receive: channel  -> perm   -> fee    -> move   -> router -> transfer

	var transferStack porttypes.IBCModule
	{
		routerMiddleware := &router.IBCMiddleware{}
		moveMiddleware := &moveibcmiddleware.IBCMiddleware{}
		feeMiddleware := &ibcfee.IBCMiddleware{}
		permMiddleware := &ibcperm.IBCMiddleware{}

		// Create Transfer Keepers
		transferKeeper := ibctransferkeeper.NewKeeper(
			appCodec,
			keys[ibctransfertypes.StoreKey],
			app.GetSubspace(ibctransfertypes.ModuleName),
			// ics4wrapper: transfer -> router
			routerMiddleware,
			app.IBCKeeper.ChannelKeeper,
			&app.IBCKeeper.PortKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			scopedTransferKeeper,
		)
		app.TransferKeeper = &transferKeeper
		transferIBCModule := ibctransfer.NewIBCModule(*app.TransferKeeper)

		// channel -> perm -> ibcfee -> move -> router -> transfer
		transferStack = permMiddleware

		// setup package forward module for multi-hop forwarding
		app.RouterKeeper = routerkeeper.NewKeeper(
			appCodec, keys[routertypes.StoreKey],
			app.GetSubspace(routertypes.ModuleName),
			app.TransferKeeper,
			app.IBCKeeper.ChannelKeeper,
			app.DistrKeeper,
			app.BankKeeper,
			// ics4wrapper: transfer -> router -> fee
			feeMiddleware,
		)
		*routerMiddleware = router.NewIBCMiddleware(
			// receive: router -> transfer
			transferIBCModule,
			app.RouterKeeper,
			0,
			routerkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
			routerkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
		)

		// create move middleware for transfer
		*moveMiddleware = moveibcmiddleware.NewIBCMiddleware(
			// receive: move -> router -> transfer
			routerMiddleware,
			// ics4wrapper: not used
			nil,
			moveKeeper,
		)

		// create ibcfee middleware for transfer
		*feeMiddleware = ibcfee.NewIBCMiddleware(
			// receive: fee -> move -> router -> transfer
			moveMiddleware,
			// ics4wrapper: transfer -> router -> fee -> channel
			*app.IBCFeeKeeper,
		)

		// create perm middleware for transfer
		*permMiddleware = ibcperm.NewIBCMiddleware(
			// receive: perm -> fee -> move -> router -> transfer
			feeMiddleware,
			// ics4wrapper: not used
			nil,
			*app.IBCPermKeeper,
		)
	}

	////////////////////////////////
	// Nft Transfer configuration //
	////////////////////////////////

	var nftTransferStack porttypes.IBCModule
	{
		// Create Transfer Keepers
		app.NftTransferKeeper = ibcnfttransferkeeper.NewKeeper(
			appCodec,
			keys[ibcnfttransfertypes.StoreKey],
			// ics4wrapper: nft transfer -> fee -> channel
			app.IBCFeeKeeper,
			app.IBCKeeper.ChannelKeeper,
			&app.IBCKeeper.PortKeeper,
			app.AccountKeeper,
			movekeeper.NewNftKeeper(moveKeeper),
			scopedNftTransferKeeper,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		)
		nftTransferIBCModule := ibcnfttransfer.NewIBCModule(*app.NftTransferKeeper)
		nftTransferStack = ibcperm.NewIBCMiddleware(
			// receive: perm -> fee -> nft transfer
			ibcfee.NewIBCMiddleware(
				// receive: channel -> fee -> nft transfer
				nftTransferIBCModule,
				*app.IBCFeeKeeper,
			),
			// ics4wrapper: not used
			nil,
			*app.IBCPermKeeper,
		)
	}

	///////////////////////
	// ICA configuration //
	///////////////////////

	var icaHostStack porttypes.IBCModule
	var icaControllerStack porttypes.IBCModule
	{
		icaHostKeeper := icahostkeeper.NewKeeper(
			appCodec, keys[icahosttypes.StoreKey],
			app.GetSubspace(icahosttypes.SubModuleName),
			app.IBCFeeKeeper,
			app.IBCKeeper.ChannelKeeper,
			&app.IBCKeeper.PortKeeper,
			app.AccountKeeper,
			scopedICAHostKeeper,
			app.MsgServiceRouter(),
		)
		app.ICAHostKeeper = &icaHostKeeper

		icaControllerKeeper := icacontrollerkeeper.NewKeeper(
			appCodec, keys[icacontrollertypes.StoreKey],
			app.GetSubspace(icacontrollertypes.SubModuleName),
			app.IBCFeeKeeper,
			app.IBCKeeper.ChannelKeeper,
			&app.IBCKeeper.PortKeeper,
			scopedICAControllerKeeper,
			app.MsgServiceRouter(),
		)
		app.ICAControllerKeeper = &icaControllerKeeper

		icaAuthKeeper := icaauthkeeper.NewKeeper(
			appCodec, keys[icaauthtypes.StoreKey],
			*app.ICAControllerKeeper,
			scopedICAAuthKeeper,
		)
		app.ICAAuthKeeper = &icaAuthKeeper

		icaAuthIBCModule := icaauth.NewIBCModule(*app.ICAAuthKeeper)
		icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)
		icaHostStack = ibcperm.NewIBCMiddleware(
			// receive: perm -> fee -> ica host
			ibcfee.NewIBCMiddleware(icaHostIBCModule, *app.IBCFeeKeeper),
			// ics4wrapper: not used
			nil,
			*app.IBCPermKeeper,
		)
		icaControllerIBCModule := icacontroller.NewIBCMiddleware(icaAuthIBCModule, *app.ICAControllerKeeper)
		icaControllerStack = ibcperm.NewIBCMiddleware(
			// receive: perm -> fee -> ica controller
			ibcfee.NewIBCMiddleware(icaControllerIBCModule, *app.IBCFeeKeeper),
			// ics4wrapper: not used
			nil,
			*app.IBCPermKeeper,
		)
	}

	//////////////////////////////
	// IBC router Configuration //
	//////////////////////////////

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icaauthtypes.ModuleName, icaControllerStack).
		AddRoute(ibcnfttransfertypes.ModuleName, nftTransferStack)

	app.IBCKeeper.SetRouter(ibcRouter)

	//////////////////////////////
	// MoveKeeper Configuration //
	//////////////////////////////

	*moveKeeper = movekeeper.NewKeeper(
		appCodec,
		keys[movetypes.StoreKey],
		app.AccountKeeper,
		app.DistrKeeper,
		// app.NftTransferKeeper,
		app.BaseApp.MsgServiceRouter(),
		moveConfig,
		// staking feature
		app.BankKeeper,
		app.DistrKeeper,
		app.StakingKeeper,
		app.RewardKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.MoveKeeper = moveKeeper
	stakingKeeper.SetSlashingHooks(app.MoveKeeper.Hooks())
	app.SetStreamingService(app.MoveKeeper.GetABCIListener())

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	// x/auction module keeper initialization

	// initialize the keeper
	auctionKeeper := auctionkeeper.NewKeeperWithRewardsAddressProvider(
		app.appCodec,
		app.keys[auctiontypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		applanes.NewRewardsAddressProvider(*app.StakingKeeper, *app.DistrKeeper),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.AuctionKeeper = &auctionKeeper

	opHostKeeper := ophostkeeper.NewKeeper(
		app.appCodec,
		app.keys[ophosttypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		ophosttypes.NewBridgeHooks(apphook.NewBridgeHook(app.IBCKeeper.ChannelKeeper, app.IBCPermKeeper)),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.OPHostKeeper = &opHostKeeper

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))

	govConfig := govtypes.DefaultConfig()
	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.MsgServiceRouter(), govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.GovKeeper = &govKeeper

	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, *app.AccountKeeper, nil, nil),
		bank.NewAppModule(appCodec, *app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, nil),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, *app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		reward.NewAppModule(appCodec, *app.RewardKeeper),
		slashing.NewAppModule(appCodec, *app.SlashingKeeper),
		distr.NewAppModule(appCodec, *app.DistrKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(*app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(*app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, *app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, *app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, *app.ConsensusParamsKeeper),
		move.NewAppModule(app.AccountKeeper, *app.MoveKeeper),
		auction.NewAppModule(app.appCodec, *app.AuctionKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		ibcnfttransfer.NewAppModule(*app.NftTransferKeeper),
		ica.NewAppModule(app.ICAControllerKeeper, app.ICAHostKeeper),
		icaauth.NewAppModule(appCodec, *app.ICAAuthKeeper),
		ibcfee.NewAppModule(*app.IBCFeeKeeper),
		router.NewAppModule(app.RouterKeeper),
		ibcperm.NewAppModule(*app.IBCPermKeeper),
		ophost.NewAppModule(*app.OPHostKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		rewardtypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		movetypes.ModuleName,
		// additional modules
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		ibcnfttransfertypes.ModuleName,
		icatypes.ModuleName,
		icaauthtypes.ModuleName,
		ibcfeetypes.ModuleName,
		routertypes.ModuleName,
		ibcpermtypes.ModuleName,
		consensusparamtypes.ModuleName,
		auctiontypes.ModuleName,
		ophosttypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName, // should be called after staking module
		slashingtypes.ModuleName,
		rewardtypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		movetypes.ModuleName,
		// additional non simd modules
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		ibcnfttransfertypes.ModuleName,
		icatypes.ModuleName,
		icaauthtypes.ModuleName,
		ibcfeetypes.ModuleName,
		routertypes.ModuleName,
		ibcpermtypes.ModuleName,
		consensusparamtypes.ModuleName,
		auctiontypes.ModuleName,
		ophosttypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		movetypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		rewardtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		feegrant.ModuleName,

		// this line is used by starport scaffolding # stargate/app/initGenesis
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		ibcnfttransfertypes.ModuleName,
		icatypes.ModuleName,
		icaauthtypes.ModuleName,
		ibcfeetypes.ModuleName,
		routertypes.ModuleName,
		ibcpermtypes.ModuleName,
		consensusparamtypes.ModuleName,
		auctiontypes.ModuleName,
		ophosttypes.ModuleName,
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// register upgrade handler for later use
	// app.RegisterUpgradeHandlers(app.configurator)

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.setPostHandler()
	app.SetEndBlocker(app.EndBlocker)

	// initialize and set the InitiaApp mempool. The current mempool will be the
	// x/auction module's mempool which will extract the top bid from the current block's auction
	// and insert the txs at the top of the block spots.
	signerExtractor := signer_extraction.NewDefaultAdapter()

	mevConfig := blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyZeroDec(),
		MaxTxs:          100,
		SignerExtractor: signerExtractor,
	}
	mevLane := mev.NewMEVLane(
		mevConfig,
		mev.NewDefaultAuctionFactory(app.txConfig.TxDecoder(), signerExtractor),
	)

	freeConfig := blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyZeroDec(),
		MaxTxs:          10,
		SignerExtractor: signerExtractor,
	}

	freeLane := freelane.NewFreeLane(
		freeConfig,
		blockbase.DefaultTxPriority(),
		applanes.FreeLaneMatchHandler(),
	)

	defaultLaneConfig := blockbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyZeroDec(),
		MaxTxs:          0,
		SignerExtractor: signerExtractor,
	}
	defaultLane := baselane.NewDefaultLane(defaultLaneConfig)

	lanes := []block.Lane{mevLane, freeLane, defaultLane}
	mempool := block.NewLanedMempool(app.Logger(), true, lanes...)
	app.SetMempool(mempool)

	anteHandler := app.setAnteHandler(mevLane, freeLane)
	for _, lane := range lanes {
		lane.SetAnteHandler(anteHandler)
	}

	// override the base-app's ABCI methods (CheckTx, PrepareProposal, ProcessProposal)
	proposalHandlers := blockabci.NewProposalHandler(
		app.Logger(),
		app.txConfig.TxDecoder(),
		app.txConfig.TxEncoder(),
		mempool,
	)

	// override base-app's ProcessProposal + PrepareProposal
	app.SetPrepareProposal(proposalHandlers.PrepareProposalHandler())
	app.SetProcessProposal(proposalHandlers.ProcessProposalHandler())

	// overrde base-app's CheckTx
	checkTxHandler := mev.NewCheckTxHandler(
		app.BaseApp,
		app.txConfig.TxDecoder(),
		mevLane,
		anteHandler,
		app.ChainID(),
	)
	app.SetCheckTx(checkTxHandler.CheckTx())

	// Load the latest state from disk if necessary, and initialize the base-app. From this point on
	// no more modifications to the base-app can be made
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// app.ScopedNftTransferKeeper = scopedNftTransferKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper
	app.ScopedICAAuthKeeper = scopedICAAuthKeeper

	return app
}

// CheckTx will check the transaction with the provided checkTxHandler. We override the default
// handler so that we can verify bid transactions before they are inserted into the mempool.
// With the POB CheckTx, we can verify the bid transaction and all of the bundled transactions
// before inserting the bid transaction into the mempool.
func (app *InitiaApp) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return app.checkTxHandler(req)
}

// SetCheckTx sets the checkTxHandler for the app.
func (app *InitiaApp) SetCheckTx(handler mev.CheckTx) {
	app.checkTxHandler = handler
}

func (app *InitiaApp) setAnteHandler(
	mevLane auctionante.MEVLane,
	freeLane block.Lane,
) sdk.AnteHandler {
	anteHandler, err := appante.NewAnteHandler(
		appante.HandlerOptions{
			HandlerOptions: cosmosante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: app.txConfig.SignModeHandler(),
				SigGasConsumer:  cosmosante.DefaultSigVerificationGasConsumer,
			},
			IBCkeeper:     app.IBCKeeper,
			MoveKeeper:    movekeeper.NewDexKeeper(app.MoveKeeper),
			Codec:         app.appCodec,
			TxEncoder:     app.txConfig.TxEncoder(),
			AuctionKeeper: *app.AuctionKeeper,
			MevLane:       mevLane,
			FreeLane:      freeLane,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	return anteHandler
}

func (app *InitiaApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *InitiaApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *InitiaApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *InitiaApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *InitiaApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *InitiaApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *InitiaApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *InitiaApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Initia's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *InitiaApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Initia's InterfaceRegistry
func (app *InitiaApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *InitiaApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *InitiaApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *InitiaApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *InitiaApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *InitiaApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// Simulate customize gas simulation to add fee deduction gas amount.
func (app *InitiaApp) Simulate(txBytes []byte) (sdk.GasInfo, *sdk.Result, error) {
	gasInfo, result, err := app.BaseApp.Simulate(txBytes)
	gasInfo.GasUsed += FeeDeductionGasAmount
	return gasInfo, result, err
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *InitiaApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(
		app.BaseApp.GRPCQueryRouter(), clientCtx,
		app.Simulate, app.interfaceRegistry,
	)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *InitiaApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx, app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry, app.Query,
	)
}

func (app *InitiaApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterUpgradeHandlers returns upgrade handlers
func (app *InitiaApp) RegisterUpgradeHandlers(cfg module.Configurator) {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		NewUpgradeHandler(app).CreateUpgradeHandler(),
	)
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) *paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(routertypes.ModuleName).WithKeyTable(routertypes.ParamKeyTable())

	return &paramsKeeper
}

// VerifyAddressLen ensures that the address matches the expected length
func VerifyAddressLen() func(addr []byte) error {
	return func(addr []byte) error {
		addrLen := len(addr)
		if addrLen != 20 && addrLen != movetypes.AddressBytesLength {
			return sdkerrors.ErrInvalidAddress
		}
		return nil
	}
}

//////////////////////////////////////
// TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *InitiaApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetAccountKeeper implements the TestingApp interface.
func (app *InitiaApp) GetAccountKeeper() *authkeeper.AccountKeeper {
	return app.AccountKeeper
}

// GetStakingKeeper implements the TestingApp interface.
func (app *InitiaApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *InitiaApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetICAControllerKeeper implements the TestingApp interface.
func (app *InitiaApp) GetICAControllerKeeper() *icacontrollerkeeper.Keeper {
	return app.ICAControllerKeeper
}

// GetICAAuthKeeper implements the TestingApp interface.
func (app *InitiaApp) GetICAAuthKeeper() *icaauthkeeper.Keeper {
	return app.ICAAuthKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *InitiaApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *InitiaApp) GetTxConfig() client.TxConfig {
	return MakeEncodingConfig().TxConfig
}

// ChainID gets chainID from private fields of BaseApp
// Should be removed once SDK 0.50.x will be adopted
func (app *InitiaApp) ChainID() string { // TODO: remove this method once chain updates to v0.50.x
	field := reflect.ValueOf(app.BaseApp).Elem().FieldByName("chainID")
	return field.String()
}