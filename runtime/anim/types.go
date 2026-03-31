package anim

// TransitionSpec describes a CSS transition for a single property.
type TransitionSpec struct {
	Property       string         // "color", "background", "all", etc.
	DurationMs     int            // transition duration in milliseconds
	TimingFunction TimingFunction // easing curve
	DelayMs        int            // delay before transition starts
}

// AnimationSpec describes a CSS keyframe animation.
type AnimationSpec struct {
	Name           string         // reference to @keyframes block
	DurationMs     int            // total cycle duration in milliseconds
	TimingFunction TimingFunction // easing per segment
	DelayMs        int            // delay before first cycle
	IterationCount int            // number of cycles (-1 = infinite)
	Direction      string         // "normal", "reverse", "alternate", "alternate-reverse"
	FillMode       string         // "none", "forwards", "backwards", "both"
	PlayState      string         // "running", "paused"
}
