package systemd

import (
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"
)

func (s *Systemd) Command(rootCmd *cobra.Command) {
	var persistentPreRunE = func(cmd *cobra.Command, args []string) error {
		_user, err := user.Current()
		if err != nil {
			return err
		}
		if _user.Gid == "0" || _user.Uid == "0" {
			return nil
		}
		return errors.New("root privileges required")
	}

	var installCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "install",
		Short:             "Systemd Install",
		Aliases:           []string{"i"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			multi, _ := cmd.Flags().GetBool("multi")
			return s.Install(multi, args...)
		},
	}

	var removeCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "remove",
		Short:             "Systemd Remove(Uninstall)",
		Aliases:           []string{"rm", "uninstall", "uni", "un"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			return s.Remove()
		},
	}
	var startCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "start [tag]...",
		Short:             "Systemd Start",
		Aliases:           []string{"run"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			num, _ := cmd.Flags().GetInt("num")
			return s.Start(num, args...)
		},
	}

	var stopCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "stop",
		Short:             "Systemd Stop",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return s.Stop(all, args...)
		},
	}
	var enableCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "enable",
		Short:             "Systemd Enable",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Enable(args...)
		},
	}

	var disableCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "disable",
		Short:             "Systemd Disable",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Disable(args...)
		},
	}

	var restartCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "restart",
		Short:             "Systemd Restart",
		Aliases:           []string{"r", "re"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return s.Restart(all, args...)
		},
	}

	var killCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "kill",
		Short:             "Systemd Kill",
		Aliases:           []string{"k"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return s.Kill(all, args...)
		},
	}

	var reloadCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "reload",
		Short:             "Systemd Reload",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return s.Reload(all, args...)
		},
	}

	var statusCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "status",
		Short:             "Systemd Status",
		Aliases:           []string{"info", "if"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := s.Status(true)
			return err
		},
	}

	var unitCmd = &cobra.Command{
		GroupID:           "systemd",
		Hidden:            true,
		Use:               "unit",
		Short:             "print systemd unit service file",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			multi, _ := cmd.Flags().GetBool("multi")
			if t, _ := cmd.Flags().GetBool("template"); t {
				execPath, err := os.Executable()
				if err != nil {
					return err
				}
				buf, err := CreateUnit(multi, s.Name, s.Description, execPath, args...)
				if err != nil {
					return err
				}
				fmt.Println(string(buf))
				return nil
			}
			fn := s.UnitFile(multi)
			s.logger.Info("filepath = " + fn)
			buf, err := os.ReadFile(fn)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
			return nil
		},
	}

	// Daemon commands
	rootCmd.AddGroup(&cobra.Group{ID: "systemd", Title: "Systemd commands"})
	rootCmd.AddCommand(
		installCmd, removeCmd, reloadCmd, unitCmd,
		startCmd, stopCmd, killCmd, restartCmd, statusCmd,
		enableCmd, disableCmd,
	)
	installCmd.Flags().BoolP("multi", "m", false, "Use template unit service")
	startCmd.Flags().IntP("num", "n", 0, "Num of Instances for start")
	stopCmd.Flags().BoolP("all", "a", false, "Stop all Instances")
	restartCmd.Flags().BoolP("all", "a", false, "Restart all Instances")
	killCmd.Flags().BoolP("all", "a", false, "Kill all Instances")
	reloadCmd.Flags().BoolP("all", "a", false, "Reload all Instances")
	unitCmd.Flags().BoolP("template", "t", false, "Show template unit service file")
	unitCmd.Flags().BoolP("multi", "m", false, "Use template unit service")
}
