package sumitest

// RunScenario executes a scenario and returns the captured frames.
func RunScenario(s Scenario) []Frame {
	app := s.NewApp(s.Width, s.Height)
	h := New(app)
	frames := make([]Frame, 0, len(s.Steps))
	for _, step := range s.Steps {
		if step.Action != nil {
			step.Action(h)
		}
		frames = append(frames, Frame{
			Name:       step.Name,
			StyledText: h.StyledText(),
		})
	}
	return frames
}
