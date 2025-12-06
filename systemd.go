package systemd

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	systemd "github.com/coreos/go-systemd/v22/dbus"
	"go.uber.org/zap"
)

type Systemd struct {
	logger      *zap.Logger
	Name        string
	Description string
	Version     string
	AppID       string

	isRoot bool
}

func New(name, desc, version, appID string) *Systemd {
	return &Systemd{
		Name:        name,
		Description: desc,
		Version:     version,
		AppID:       appID,
		logger:      zap.L(),
	}
}

func (s *Systemd) WithLogger(logger *zap.Logger) *Systemd {
	s.logger = logger
	return s
}

func (s *Systemd) UnitFilePath() string {
	if s.isRoot {
		return filepath.Join("/", "etc", "systemd", "system", s.Name+".service")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "systemd", "system", s.Name+".service")
}

func (s *Systemd) Install(args ...string) error {
	s.logger.Info("Install... " + s.Name)
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	var buf []byte
	buf, err = CreateUnit(s.Name, s.Description, execPath, args...)
	if err != nil {
		return err
	}
	path := s.UnitFilePath()
	err = os.WriteFile(path, buf, 0644)
	if err != nil {
		s.logger.Error("Failed to write unit file", zap.String("path", path), zap.Error(err))
		return err
	}
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	s.logger.Info("Installed " + s.Name)
	return conn.ReloadContext(ctx)
}

// Remove the service
func (s *Systemd) Remove() error {
	s.logger.Info("Removing... " + s.Name)
	err := s.Stop()
	if err != nil {
		s.logger.Warn(err.Error())
	}
	err = s.Disable()
	if err != nil {
		s.logger.Warn(err.Error())
	}
	err = os.Remove(s.UnitFilePath())
	if err != nil {
		return errors.New("remove failed")
	}
	s.logger.Info("Removed " + s.Name)
	return nil
}

// Start the service
func (s *Systemd) Start() error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	name := s.Name + ".service"
	_, err = conn.StartUnitContext(ctx, name, "fail", recv)
	if err != nil {
		return err
	}
	v := <-recv
	if v == "failed" {
		s.logger.Error("Started [ " + name + " ] " + v)
	} else {
		s.logger.Info("Started [ " + name + " ] " + v)
	}
	return nil
}

// Stop the service
func (s *Systemd) Stop() error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	name := s.Name + ".service"
	_, err = conn.StopUnitContext(ctx, name, "fail", recv)
	if err != nil {
		return err
	}
	v := <-recv
	if v == "failed" {
		s.logger.Error("Stop [" + name + "] " + v)
	} else {
		s.logger.Info("Stop [ " + name + " ] " + v)
	}
	return nil
}

func fileExists(name string) bool {
	fi, err := os.Stat(name)
	return err == nil && !fi.IsDir()
}

// Enable the service
func (s *Systemd) Enable() (err error) {
	origin := s.UnitFilePath()
	if fileExists(origin) {
		target := filepath.Join("/", "etc", "systemd", "system", "multi-user.target.wants", s.Name+".service")
		err = os.Symlink(origin, target)
		if err != nil {
			s.logger.Error("Failed to create symlink", zap.String("origin", origin), zap.String("target", target), zap.Error(err))
		} else {
			s.logger.Info("Created symlink", zap.String("target", target), zap.String("origin", origin))
		}
		return nil
	}
	return errors.New("service is not installed")
}

// Disable the service
func (s *Systemd) Disable() (err error) {
	target := filepath.Join("/", "etc", "systemd", "system", "multi-user.target.wants", s.Name+".service")
	err = os.Remove(target)
	if err != nil {
		s.logger.Warn("Failed to remove symlink", zap.String("target", target), zap.Error(err))
	}
	return nil
}

// Kill the service
func (s *Systemd) Kill() error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	conn.KillUnitWithTarget(ctx, s.Name+".service", systemd.All, 9)
	return nil
}

// Restart the service
func (s *Systemd) Restart() error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	name := s.Name + ".service"
	_, err = conn.RestartUnitContext(ctx, name, "fail", recv)
	if err != nil {
		return err
	}
	v := <-recv
	if v == "failed" {
		s.logger.Error("Restarted [ " + name + " ] " + v)
	} else {
		s.logger.Info("Restarted [ " + name + " ] " + v)
	}
	return nil
}

// Reload the service
func (s *Systemd) Reload() error {
	s.logger.Info("Reloading... " + s.Name)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	name := s.Name + ".service"
	_, err = conn.ReloadOrRestartUnitContext(ctx, name, "fail", recv)
	if err != nil {
		return err
	}
	v := <-recv
	if v == "failed" {
		s.logger.Error("Reloaded [ " + name + " ] " + v)
	} else {
		s.logger.Info("Reloaded [ " + name + " ] " + v)
	}
	return nil
}

// Status - Get service status
func (s *Systemd) Status(show bool) ([]systemd.UnitStatus, error) {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := conn.ListUnitsByPatternsContext(ctx, nil, []string{s.Name + "*"})
	if err != nil {
		return nil, err
	}
	if show {
		for _, item := range items {
			if item.SubState == "running" {
				s.logger.Info("Status", zap.String("name", item.Name), zap.String("active", item.ActiveState), zap.String("sub", item.SubState))
			} else {
				s.logger.Warn("Status", zap.String("name", item.Name), zap.String("active", item.ActiveState), zap.String("sub", item.SubState))
			}
		}
	}
	return items, nil
}
