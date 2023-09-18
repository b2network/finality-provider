package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/babylonchain/babylon/x/checkpointing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/urfave/cli"

	"github.com/babylonchain/btc-validator/proto"
	dc "github.com/babylonchain/btc-validator/service/client"
	"github.com/babylonchain/btc-validator/valcfg"
)

var daemonCommands = []cli.Command{
	{
		Name:      "daemon",
		ShortName: "dn",
		Usage:     "More advanced commands which require validator daemon to be running.",
		Category:  "Daemon commands",
		Subcommands: []cli.Command{
			getDaemonInfoCmd,
			createValDaemonCmd,
			lsValDaemonCmd,
			valInfoDaemonCmd,
			registerValDaemonCmd,
			addFinalitySigDaemonCmd,
		},
	},
}

const (
	valdDaemonAddressFlag = "daemon-address"
	keyNameFlag           = "key-name"
	valBabylonPkFlag      = "babylon-pk"
	blockHeightFlag       = "height"
	lastCommitHashFlag    = "last-commit-hash"

	// flags for description
	monikerFlag          = "moniker"
	identityFlag         = "identity"
	websiteFlag          = "website"
	securityContractFlag = "security-contract"
	detailsFlag          = "details"
)

var (
	defaultValdDaemonAddress = "127.0.0.1:" + strconv.Itoa(valcfg.DefaultRPCPort)
	defaultLastCommitHashStr = "fd903d9baeb3ab1c734ee003de75f676c5a9a8d0574647e5385834d57d3e79ec"
)

var getDaemonInfoCmd = cli.Command{
	Name:      "get-info",
	ShortName: "gi",
	Usage:     "Get information of the running daemon.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
	},
	Action: getInfo,
}

func getInfo(ctx *cli.Context) error {
	daemonAddress := ctx.String(valdDaemonAddressFlag)
	client, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	info, err := client.GetInfo(context.Background())

	if err != nil {
		return err
	}

	printRespJSON(info)

	return nil
}

var createValDaemonCmd = cli.Command{
	Name:      "create-validator",
	ShortName: "cv",
	Usage:     "Create a Bitcoin validator object and save it in database.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
		cli.StringFlag{
			Name:     keyNameFlag,
			Usage:    "The unique name of the validator key",
			Required: true,
		},
		cli.StringFlag{
			Name:  monikerFlag,
			Usage: "A human-readable name for the validator",
			Value: "",
		},
		cli.StringFlag{
			Name:  identityFlag,
			Usage: "An optional identity signature (ex. UPort or Keybase)",
			Value: "",
		},
		cli.StringFlag{
			Name:  websiteFlag,
			Usage: "An optional website link",
			Value: "",
		},
		cli.StringFlag{
			Name:  securityContractFlag,
			Usage: "An optional email for security contact",
			Value: "",
		},
		cli.StringFlag{
			Name:  detailsFlag,
			Usage: "Other optional details",
			Value: "",
		},
	},
	Action: createValDaemon,
}

func createValDaemon(ctx *cli.Context) error {
	daemonAddress := ctx.String(valdDaemonAddressFlag)
	keyName := ctx.String(keyNameFlag)
	description, err := getDesciptionFromContext(ctx)
	if err != nil {
		return err
	}

	client, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	info, err := client.CreateValidator(context.Background(), keyName, &description)

	if err != nil {
		return err
	}

	printRespJSON(info)

	return nil
}

func getDesciptionFromContext(ctx *cli.Context) (stakingtypes.Description, error) {

	// get information for description
	monikerStr := ctx.String(monikerFlag)
	identityStr := ctx.String(identityFlag)
	websiteStr := ctx.String(websiteFlag)
	securityContractStr := ctx.String(securityContractFlag)
	detailsStr := ctx.String(detailsFlag)

	description := stakingtypes.NewDescription(monikerStr, identityStr, websiteStr, securityContractStr, detailsStr)
	return description.EnsureLength()
}

var lsValDaemonCmd = cli.Command{
	Name:      "list-validators",
	ShortName: "ls",
	Usage:     "List validators stored in the database.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
	},
	Action: lsValDaemon,
}

func lsValDaemon(ctx *cli.Context) error {
	daemonAddress := ctx.String(valdDaemonAddressFlag)
	rpcClient, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	resp, err := rpcClient.QueryValidatorList(context.Background())
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

var valInfoDaemonCmd = cli.Command{
	Name:      "validator-info",
	ShortName: "vi",
	Usage:     "Show the information of the validator.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
		cli.StringFlag{
			Name:     valBabylonPkFlag,
			Usage:    "The hex string of the Babylon public key",
			Required: true,
		},
	},
	Action: valInfoDaemon,
}

func valInfoDaemon(ctx *cli.Context) error {
	daemonAddress := ctx.String(valdDaemonAddressFlag)
	rpcClient, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	bbnPkBytes, err := hex.DecodeString(ctx.String(valBabylonPkFlag))
	if err != nil {
		return err
	}

	resp, err := rpcClient.QueryValidatorInfo(context.Background(), bbnPkBytes)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}

var registerValDaemonCmd = cli.Command{
	Name:      "register-validator",
	ShortName: "rv",
	Usage:     "Register a created Bitcoin validator to Babylon.",
	UsageText: fmt.Sprintf("register-validator --%s [key-name]", keyNameFlag),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
		cli.StringFlag{
			Name:     keyNameFlag,
			Usage:    "The unique name of the validator key",
			Required: true,
		},
	},
	Action: registerVal,
}

func registerVal(ctx *cli.Context) error {
	keyName := ctx.String(keyNameFlag)

	daemonAddress := ctx.String(valdDaemonAddressFlag)
	rpcClient, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	res, err := rpcClient.RegisterValidator(context.Background(), keyName)
	if err != nil {
		return err
	}

	printRespJSON(res)

	return nil
}

// addFinalitySigDaemonCmd allows manual submission of finality signatures
// NOTE: should only be used for presentation/testing purposes
var addFinalitySigDaemonCmd = cli.Command{
	Name:      "add-finality-sig",
	ShortName: "afs",
	Usage:     "Send a finality signature to Babylon. This command should only be used for presentation/testing purposes",
	UsageText: fmt.Sprintf("add-finality-sig --%s [babylon_pk_hex]", valBabylonPkFlag),
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  valdDaemonAddressFlag,
			Usage: "Full address of the validator daemon in format tcp://<host>:<port>",
			Value: defaultValdDaemonAddress,
		},
		cli.StringFlag{
			Name:     valBabylonPkFlag,
			Usage:    "The hex string of the Babylon public key",
			Required: true,
		},
		cli.Uint64Flag{
			Name:     blockHeightFlag,
			Usage:    "The height of the Babylon block",
			Required: true,
		},
		cli.StringFlag{
			Name:  lastCommitHashFlag,
			Usage: "The last commit hash of the Babylon block",
			Value: defaultLastCommitHashStr,
		},
	},
	Action: addFinalitySig,
}

func addFinalitySig(ctx *cli.Context) error {
	daemonAddress := ctx.String(valdDaemonAddressFlag)
	rpcClient, cleanUp, err := dc.NewValidatorServiceGRpcClient(daemonAddress)
	if err != nil {
		return err
	}
	defer cleanUp()

	bbnPk, err := proto.NewBabylonPkFromHex(ctx.String(valBabylonPkFlag))
	if err != nil {
		return err
	}

	lch, err := types.NewLastCommitHashFromHex(ctx.String(lastCommitHashFlag))
	if err != nil {
		return err
	}

	res, err := rpcClient.AddFinalitySignature(
		context.Background(), bbnPk.Key, ctx.Uint64(blockHeightFlag), lch)
	if err != nil {
		return err
	}

	printRespJSON(res)

	return nil
}
