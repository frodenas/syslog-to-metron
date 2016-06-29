package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/logs"
	"github.com/pivotal-golang/lager"
	"gopkg.in/mcuadros/go-syslog.v2"
)

var (
	debug          bool
	metronAddress  string
	metronOrigin   string
	syslogAddress  string
	syslogProtocol string
	syslogFormat   string
	sourceType     string
	sourceInstance string
	appIDs         appID
)

type appID []string

func (a *appID) String() string {
	return fmt.Sprint(appIDs)
}

func (m *appID) Set(value string) error {
	if appIDs == nil {
		appIDs = appID{}
	}

	appIDs = append(appIDs, value)

	return nil
}

func main() {
	flag.BoolVar(&debug, "debug", false, "Output debug logging")
	flag.StringVar(&metronAddress, "metron-address", "127.0.0.1:3457", "Metron address (e.g. 127.0.0.1:3457)")
	flag.StringVar(&metronOrigin, "metron-origin", "", "Source name for logs emitted by this process (e.g. redis)")
	flag.StringVar(&syslogAddress, "syslog-address", "127.0.0.1:10514", "Syslog Listen Address (e.g. 127.0.0.1:10514)")
	flag.StringVar(&syslogProtocol, "syslog-protocol", "UDP", "Syslog Protocol (TCP|UDP|Unix)")
	flag.StringVar(&syslogFormat, "syslog-format", "Automatic", "Syslog Format (RFC3164,RFC5424,RFC6587,Automatic)")
	flag.StringVar(&sourceType, "source-type", "SRV", "Logs Source Type")
	flag.StringVar(&sourceInstance, "source-instance", "0", "Logs Source Instance")
	flag.Var(&appIDs, "app-id", "Logs Application ID (it can be specified multiple times)")
	flag.Parse()

	stdoutLogLevel := lager.INFO
	if debug {
		stdoutLogLevel = lager.DEBUG
	}

	logger := lager.NewLogger("syslog-to-metron")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, stdoutLogLevel))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Debug("dropsonde", lager.Data{"metron-address": metronAddress, "metron-origin": metronOrigin})
	if err := dropsonde.Initialize(metronAddress, metronOrigin); err != nil {
		logger.Error("dropsonde", fmt.Errorf("Dropsonde failed to initialize: '%s'", err))
		os.Exit(1)
	}

	logger.Debug("syslog-listener", lager.Data{"syslog-address": syslogAddress, "syslog-protocol": syslogProtocol, "syslog-format": syslogFormat})
	logChannel := make(syslog.LogPartsChannel)
	logHandler := syslog.NewChannelHandler(logChannel)
	server := syslog.NewServer()
	server.SetHandler(logHandler)

	switch syslogFormat {
	case "RFC3164":
		server.SetFormat(syslog.RFC3164)
	case "RFC5424":
		server.SetFormat(syslog.RFC5424)
	case "RFC6587":
		server.SetFormat(syslog.RFC6587)
	case "Automatic":
		server.SetFormat(syslog.Automatic)
	default:
		logger.Error("syslog-listener", fmt.Errorf("Syslog Format '%s' is not supported", syslogFormat))
		os.Exit(1)
	}

	switch syslogProtocol {
	case "TCP":
		if err := server.ListenTCP(syslogAddress); err != nil {
			logger.Error("syslog-listener", fmt.Errorf("Syslog TCP listener failed to initialize: '%s'", err))
			os.Exit(1)
		}
	case "UDP":
		if err := server.ListenUDP(syslogAddress); err != nil {
			logger.Error("syslog-listener", fmt.Errorf("Syslog UDP listener failed to initialize: '%s'", err))
			os.Exit(1)
		}
	case "Unix":
		if err := server.ListenUnixgram(syslogAddress); err != nil {
			logger.Error("syslog-listener", fmt.Errorf("Syslog Unix listener failed to initialize: '%s'", err))
			os.Exit(1)
		}
	default:
		logger.Error("syslog-listener", fmt.Errorf("Syslog Protocol '%s' is not supported", syslogProtocol))
		os.Exit(1)
	}

	if err := server.Boot(); err != nil {
		logger.Error("syslog-listener", fmt.Errorf("Syslog listener failed to boot: '%s'", err))
		os.Exit(1)
	}

	go func(logChannel syslog.LogPartsChannel) {
		var message string

		for logParts := range logChannel {
			if logParts["message"] != nil {
				message = logParts["message"].(string)
			} else if logParts["content"] != nil {
				message = logParts["content"].(string)
			} else {
				continue
			}

			for _, appID := range appIDs {
				logger.Debug("dropsonde", lager.Data{"appID": appID, "message": message, "sourceType": sourceType, "sourceInstance": sourceInstance})
				if err := logs.SendAppLog(appID, message, sourceType, sourceInstance); err != nil {
					logger.Error("dropsonde", fmt.Errorf("Error sending logs to Dropsonde: '%s'", err))
				}
			}
		}
	}(logChannel)

	server.Wait()
}
