package mvm

type Shell struct {
	parent  *Shell
	frame   *Frame
	execute bool
	running bool
	object  Object
}

func MakeShell(frame *Frame, parent *Shell) *Shell {
	s := &Shell{
		frame:  frame,
		parent: parent,
	}
	if parent != nil {
		m := parent.object.(*Machine)
		m.shells[frame] = s
	}
	return s
}

func (s *Shell) MarkForExecution() {
	s.execute = true
	tasks <- s
}
