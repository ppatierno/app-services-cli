package main

import (
	"context"
	"fmt"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/debug"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/root"
	"os"
	"strings"

	"github.com/redhat-developer/app-services-cli/internal/telemetry"
	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/factory"
	"github.com/redhat-developer/app-services-cli/pkg/core/cmdutil/factory/defaultfactory"
	"github.com/redhat-developer/app-services-cli/pkg/core/config"
	"github.com/redhat-developer/app-services-cli/pkg/core/ioutil/icon"
	"github.com/redhat-developer/app-services-cli/pkg/core/localize"
	"github.com/redhat-developer/app-services-cli/pkg/core/localize/goi18n"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/app-services-cli/internal/build"
)

func main() {
	localizer, err := goi18n.New(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	buildVersion := build.Version
	cmdFactory := defaultfactory.New(localizer)

	err = initConfig(cmdFactory)
	if err != nil {
		cmdFactory.Logger.Errorf(localizer.MustLocalize("main.config.error", localize.NewEntry("Error", err)))
		os.Exit(1)
	}

	rootCmd := root.NewRootCommand(cmdFactory, buildVersion)
	rootCmd.InitDefaultHelpCmd()

	err = executeCommandWithTelemetry(rootCmd, cmdFactory)

	if err == nil {
		if debug.Enabled() {
			build.CheckForUpdate(cmdFactory.Context, build.Version, cmdFactory.Logger, localizer)
		}
		return
	}
	cmdFactory.Logger.Errorf("%v\n", rootError(err, localizer))
	build.CheckForUpdate(context.Background(), build.Version, cmdFactory.Logger, localizer)
	os.Exit(1)
}

func initConfig(f *factory.Factory) error {
	if !config.HasCustomLocation() {
		rhoasCfgDir, err := config.DefaultDir()
		if err != nil {
			return err
		}

		// create rhoas config directory
		if _, err = os.Stat(rhoasCfgDir); os.IsNotExist(err) {
			err = os.MkdirAll(rhoasCfgDir, 0o700)
			if err != nil {
				return err
			}
		}
	}

	cfgFile, err := f.Config.Load()

	if cfgFile != nil {
		return err
	}

	if !os.IsNotExist(err) {
		return err
	}

	cfgFile = &config.Config{}
	if err := f.Config.Save(cfgFile); err != nil {
		return err
	}
	return nil
}

// rootError creates the root error which is printed to the console
// it wraps the error which has been returned from subcommands with a prefix
func rootError(err error, localizer localize.Localizer) error {
	prefix := icon.ErrorPrefix()
	errMessage := err.Error()
	if prefix == icon.ErrorSymbol {
		errMessage = firstCharToUpper(errMessage)
	}

	if strings.Contains(errMessage, "\n") {
		return fmt.Errorf("%v %v\n%v", icon.ErrorPrefix(), errMessage, localizer.MustLocalize("common.log.error.verboseModeHint"))
	}
	return fmt.Errorf("%v %v. %v", icon.ErrorPrefix(), errMessage, localizer.MustLocalize("common.log.error.verboseModeHint"))
}

// Ensure that the first character in the string is capitalized
func firstCharToUpper(message string) string {
	return strings.ToUpper(message[:1]) + message[1:]
}

func executeCommandWithTelemetry(rootCmd *cobra.Command, cmdFactory *factory.Factory) error {
	telemetry, err := telemetry.CreateTelemetry(cmdFactory)
	if err != nil {
		cmdFactory.Logger.Errorf(cmdFactory.Localizer.MustLocalize("main.config.error", localize.NewEntry("Error", err)))
		os.Exit(1)
	}
	commandPath := ""
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cmdFactory.Logger.SetDebug(debug.Enabled())
		if cmd.Runnable() && !cmd.Hidden {
			commandPath = cmd.CommandPath()
		}
	}
	err = rootCmd.Execute()

	if commandPath != "" {
		telemetry.Finish(commandPath, err)
	}
	return err
}
