// Copyright © 2020 Ettore Di Giacinto <mudler@mocaccino.org>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, see <http://www.gnu.org/licenses/>.
package cmd

import (
	"github.com/mudler/luet/pkg/api/core/types"
	installer "github.com/mudler/luet/pkg/installer"
	"github.com/mudler/luet/pkg/solver"

	helpers "github.com/mudler/luet/cmd/helpers"
	"github.com/mudler/luet/cmd/util"

	pkg "github.com/mudler/luet/pkg/package"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var replaceCmd = &cobra.Command{
	Use:     "replace <pkg1> <pkg2> --for <pkg3> --for <pkg4> ...",
	Short:   "replace a set of packages",
	Aliases: []string{"r"},
	Long: `Replaces one or a group of packages without asking questions:

	$ luet replace -y system/busybox ... --for shells/bash --for system/coreutils ...
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		util.BindSystemFlags(cmd)
		util.BindSolverFlags(cmd)
		viper.BindPFlag("onlydeps", cmd.Flags().Lookup("onlydeps"))
		viper.BindPFlag("nodeps", cmd.Flags().Lookup("nodeps"))
		viper.BindPFlag("force", cmd.Flags().Lookup("force"))
		viper.BindPFlag("for", cmd.Flags().Lookup("for"))

		viper.BindPFlag("yes", cmd.Flags().Lookup("yes"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var toUninstall pkg.Packages
		var toAdd pkg.Packages

		f := viper.GetStringSlice("for")
		force := viper.GetBool("force")
		nodeps := viper.GetBool("nodeps")
		onlydeps := viper.GetBool("onlydeps")
		yes := viper.GetBool("yes")
		downloadOnly, _ := cmd.Flags().GetBool("download-only")

		util.SetSystemConfig(util.DefaultContext)
		util.SetSolverConfig(util.DefaultContext)
		for _, a := range args {
			pack, err := helpers.ParsePackageStr(a)
			if err != nil {
				util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
			}
			toUninstall = append(toUninstall, pack)
		}

		for _, a := range f {
			pack, err := helpers.ParsePackageStr(a)
			if err != nil {
				util.DefaultContext.Fatal("Invalid package string ", a, ": ", err.Error())
			}
			toAdd = append(toAdd, pack)
		}

		util.DefaultContext.Config.GetSolverOptions().Implementation = solver.SingleCoreSimple

		util.DefaultContext.Debug("Solver", util.DefaultContext.Config.GetSolverOptions().CompactString())

		// Load config protect configs
		util.DefaultContext.Config.LoadConfigProtect(util.DefaultContext)

		inst := installer.NewLuetInstaller(installer.LuetInstallerOptions{
			Concurrency:                 util.DefaultContext.Config.GetGeneral().Concurrency,
			SolverOptions:               *util.DefaultContext.Config.GetSolverOptions(),
			NoDeps:                      nodeps,
			Force:                       force,
			OnlyDeps:                    onlydeps,
			PreserveSystemEssentialData: true,
			Ask:                         !yes,
			DownloadOnly:                downloadOnly,
			PackageRepositories:         util.DefaultContext.Config.SystemRepositories,
			Context:                     util.DefaultContext,
		})

		system := &installer.System{Database: util.DefaultContext.Config.GetSystemDB(), Target: util.DefaultContext.Config.GetSystem().Rootfs}
		err := inst.Swap(toUninstall, toAdd, system)
		if err != nil {
			util.DefaultContext.Fatal("Error: " + err.Error())
		}
	},
}

func init() {

	replaceCmd.Flags().String("system-dbpath", "", "System db path")
	replaceCmd.Flags().String("system-target", "", "System rootpath")
	replaceCmd.Flags().String("system-engine", "", "System DB engine")

	replaceCmd.Flags().String("solver-type", "", "Solver strategy ( Defaults none, available: "+types.AvailableResolvers+" )")
	replaceCmd.Flags().Float32("solver-rate", 0.7, "Solver learning rate")
	replaceCmd.Flags().Float32("solver-discount", 1.0, "Solver discount rate")
	replaceCmd.Flags().Int("solver-attempts", 9000, "Solver maximum attempts")
	replaceCmd.Flags().Bool("nodeps", false, "Don't consider package dependencies (harmful!)")
	replaceCmd.Flags().Bool("onlydeps", false, "Consider **only** package dependencies")
	replaceCmd.Flags().Bool("force", false, "Skip errors and keep going (potentially harmful)")
	replaceCmd.Flags().Bool("solver-concurrent", false, "Use concurrent solver (experimental)")
	replaceCmd.Flags().BoolP("yes", "y", false, "Don't ask questions")
	replaceCmd.Flags().StringSlice("for", []string{}, "Packages that has to be installed in place of others")
	replaceCmd.Flags().Bool("download-only", false, "Download only")

	RootCmd.AddCommand(replaceCmd)
}
