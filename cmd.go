package systemd

import (
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
		s.isRoot = _user.Gid == "0" || _user.Uid == "0"
		return nil
	}

	var installCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "install",
		Short:             "Systemd Install",
		Aliases:           []string{"i"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Install()
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
			return s.Start()
		},
	}

	var stopCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "stop",
		Short:             "Systemd Stop",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Stop()
		},
	}
	var enableCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "enable",
		Short:             "Systemd Enable",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Enable()
		},
	}

	var disableCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "disable",
		Short:             "Systemd Disable",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Disable()
		},
	}

	var restartCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "restart",
		Short:             "Systemd Restart",
		Aliases:           []string{"r", "re"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Restart()
		},
	}

	var killCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "kill",
		Short:             "Systemd Kill",
		Aliases:           []string{"k"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Kill()
		},
	}

	var reloadCmd = &cobra.Command{
		GroupID:           "systemd",
		Use:               "reload",
		Short:             "Systemd Reload",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.Reload()
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
			if t, _ := cmd.Flags().GetBool("template"); t {
				execPath, err := os.Executable()
				if err != nil {
					return err
				}
				buf, err := CreateUnit(s.Name, s.Description, execPath)
				if err != nil {
					return err
				}
				fmt.Println(string(buf))
				return nil
			}
			fn := s.UnitFilePath()
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
	unitCmd.Flags().BoolP("template", "t", false, "Show template unit service file")
}
