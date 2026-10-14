package cli

import "errors"

// handleBuild は路線を敷設する。
func handleBuild(s *Session, cmd *Command) error {
	from, ok := cmd.Arg(0)
	if !ok {
		return errors.New("出発都市を指定してください: build <from> <to> [--gauge N] [--double]")
	}
	to, ok := cmd.Arg(1)
	if !ok {
		return errors.New("到着都市を指定してください: build <from> <to> [--gauge N] [--double]")
	}
	gauge, err := cmd.OptInt("gauge", 1067)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.BuildLine(ctx, from, to, gauge, cmd.Flag("double"))
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}

// handleBuildFacility は開発施設を建設する。
func handleBuildFacility(s *Session, cmd *Command) error {
	city, ok := cmd.Arg(0)
	if !ok {
		return errors.New("都市を指定してください: build-facility <city> <kind> [--scale N]")
	}
	kind, ok := cmd.Arg(1)
	if !ok {
		return errors.New("施設種別を指定してください: build-facility <city> <kind> [--scale N]")
	}
	scale, err := cmd.OptInt("scale", 1)
	if err != nil {
		return err
	}
	ctx, cancel := s.ctx()
	defer cancel()

	events, err := s.Game.BuildFacility(ctx, city, kind, scale)
	if err != nil {
		return err
	}
	s.printEvents(events)
	return nil
}
