package ui

import (
	"github.com/mafik/mvm/vec2"
)

type Widget interface {
	Options(vec2.Vec2) []Option
}

type Action interface {
	Move(TouchContext) Action
	End(TouchContext)
}

type PreMoveAction interface {
	PreMove(TouchContext)
}

type Option interface {
	Name() string
	Keycode() string
	Activate(TouchContext) Action
}

type OptionContext struct {
	Path    TreePath
	Options []Option
}

func QueryOptions(travCtx TraversalContext, root interface{}, v vec2.Vec2) []OptionContext {
	result := []OptionContext{}
	WalkAtPoint(travCtx, root, v, func(path TreePath, v vec2.Vec2) WalkAction {
		last := path[len(path)-1]
		opts := OptionContext{path, nil}
		if editable, ok := last.(Editable); ok {
			opts.Options = append(opts.Options, EditOption{travCtx, editable})
		}
		if widget, ok := last.(Widget); ok {
			opts.Options = append(opts.Options, widget.Options(v)...)
		}
		if len(opts.Options) > 0 {
			result = append(result, opts)
		}
		return Explore
	})
	return result
}

func PickOption(options []OptionContext, keycode string, key string) (TreePath, Option) {
	for _, optCtx := range options {
		for _, opt := range optCtx.Options {
			if editOpt, ok := opt.(EditOption); ok && editKeys[keycode] {
				if editOpt.IsEditing() {
					return optCtx.Path, TypeOption{editOpt.Editable, keycode, key}
				}
			}
		}
	}
	for i, _ := range options {
		optCtx := options[len(options)-1-i]
		for _, opt := range optCtx.Options {
			if opt.Keycode() == keycode {
				return optCtx.Path, opt
			}
		}
	}
	return nil, nil
}
