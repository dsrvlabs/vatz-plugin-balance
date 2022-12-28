package main

import (
	"flag"
	"fmt"
	"math"
	"strconv"
	"time"

	rpcBalances "github.com/dsrvlabs/vatz-plugin-balance/rpc/cosmos"
	pluginpb "github.com/dsrvlabs/vatz-proto/plugin/v1"
	"github.com/dsrvlabs/vatz/sdk"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	// Default values.
	defaultAddr        = "127.0.0.1"
	defaultPort        = 9094
	defaultApiPort     = 1317
	defaultAccountAddr = "address"

	pluginName = "node-balance-alarm"
)

var (
	addr              string
	port              int
	apiPort           uint
	accountAddr       string
	firstRun          int
	checkPointBalance float64
	severity          = pluginpb.SEVERITY_INFO
	startTime         time.Time
)

func init() {
	flag.StringVar(&addr, "addr", defaultAddr, "IP Address(e.g. 0.0.0.0, 127.0.0.1)")
	flag.IntVar(&port, "port", defaultPort, "Port number, default 9094")
	flag.UintVar(&apiPort, "apiPort", defaultApiPort, "Need to know API port")
	flag.StringVar(&accountAddr, "accountAddr", defaultAccountAddr, "Need account address")

	flag.Parse()

	firstRun = 0
	startTime = time.Now()
}

func main() {
	p := sdk.NewPlugin(pluginName)
	p.Register(pluginFeature)

	ctx := context.Background()
	if err := p.Start(ctx, addr, port); err != nil {
		fmt.Println("exit")
	}
}

func pluginFeature(info, option map[string]*structpb.Value) (sdk.CallResponse, error) {
	//severity := pluginpb.SEVERITY_INFO
	state := pluginpb.STATE_NONE

	var msg string
	accntInfo, _ := rpcBalances.GetAccountInfo(apiPort, accountAddr)
	balance, _ := rpcBalances.GetBalances(apiPort, accountAddr)
	bal, _ := strconv.Atoi(balance)
	fbalance := float64(bal) * math.Pow(0.1, 18)
	if firstRun == 0 {
		firstRun++
		msg += "First check balances!\n"
		msg += accntInfo + "\n"
		msg += fmt.Sprintf("Amount = %f", fbalance)
		checkPointBalance = fbalance
		log.Info().Str("module", "plugin").Msg(msg)
		if severity == pluginpb.SEVERITY_INFO {
			severity = pluginpb.SEVERITY_WARNING
		} else {
			severity = pluginpb.SEVERITY_INFO
		}
		state = pluginpb.STATE_SUCCESS
	} else {
		if checkPointBalance > fbalance {
			msg += "Balances changed.\n"
			msg += accntInfo + "\n"
			msg += fmt.Sprintf("Amount = %f", fbalance)
			log.Info().Str("module", "plugin").Msg(msg)
			if time.Now().After(startTime.Add(time.Hour * 12)) {
				if severity == pluginpb.SEVERITY_INFO {
					severity = pluginpb.SEVERITY_WARNING
				} else {
					severity = pluginpb.SEVERITY_INFO
				}
				state = pluginpb.STATE_SUCCESS
			}
		} else if checkPointBalance == fbalance {
			msg += "Latest balances check.\n"
			msg += accntInfo + "\n"
			msg += fmt.Sprintf("Amount = %f", fbalance)
			log.Info().Str("module", "plugin").Msg(msg)
			fmt.Println("Now: ", time.Now())
			fmt.Println("After 12 Hour: ", startTime.Add(time.Hour*12))
			if time.Now().After(startTime.Add(time.Hour*12)) || firstRun == 1 {
				severity = pluginpb.SEVERITY_INFO
				state = pluginpb.STATE_SUCCESS
				startTime = time.Now()
				fmt.Println("Time expired! Reset startTime = ", startTime)
				fmt.Println("firstRun = ", firstRun)
				if firstRun == 1 {
					firstRun++
				}
			}
		}

		checkPointBalance = fbalance
	}

	if checkPointBalance < 0.001000 {
		//Maybe this account will be empty in some hours
		msg = ""
		msg += "A deposit is required.\n"
		msg += fmt.Sprintf("Amount = %f", fbalance)
		log.Warn().Str("module", "plugin").Msg(msg)
		if severity == pluginpb.SEVERITY_INFO {
			severity = pluginpb.SEVERITY_WARNING
		} else {
			severity = pluginpb.SEVERITY_INFO
		}
		state = pluginpb.STATE_SUCCESS
	}
	ret := sdk.CallResponse{
		FuncName:   info["execute_method"].GetStringValue(),
		Message:    msg,
		Severity:   severity,
		State:      state,
		AlertTypes: []pluginpb.ALERT_TYPE{pluginpb.ALERT_TYPE_DISCORD},
	}
	return ret, nil
}
