package ui

import (
	"github.com/mafik/mvm/vec2"
)

type actionContext struct {
	path   TreePath
	action Action
}

type Touch struct {
	Curr    vec2.Vec2
	Last    vec2.Vec2
	Wheel   float64
	actions map[string]actionContext
}

func MakeTouch() Touch {
	return Touch{actions: make(map[string]actionContext)}
}

type TouchContext struct {
	textMeasurer TextMeasurer
	Touch        *Touch
	Path         TreePath
}

func (ctx TouchContext) At(check func(interface{}) bool) TouchContext {
	for i := len(ctx.Path) - 1; i >= 0; i -= 1 {
		if check(ctx.Path[i]) {
			return TouchContext{ctx.textMeasurer, ctx.Touch, ctx.Path[:i+1]}
		}
	}
	return TouchContext{ctx.textMeasurer, ctx.Touch, ctx.Path[:1]}
}
func (ctx TouchContext) AtTopBlueprint() TouchContext {
	return TouchContext{ctx.textMeasurer, ctx.Touch, ctx.Path[:2]}
}
func (ctx TouchContext) Position() vec2.Vec2 {
	return ToLocal(ctx.textMeasurer, ctx.Path, ctx.Touch.Curr)
}
func (ctx TouchContext) Delta() vec2.Vec2 {
	curr := ToLocal(ctx.textMeasurer, ctx.Path, ctx.Touch.Curr)
	last := ToLocal(ctx.textMeasurer, ctx.Path, ctx.Touch.Last)
	return vec2.Sub(curr, last)
}
func (ctx TouchContext) Query(callback func(TreePath, vec2.Vec2) WalkAction) {
	pos := ToLocal(ctx.textMeasurer, ctx.Path[:len(ctx.Path)-1], ctx.Touch.Curr)
	WalkAtPoint(ctx.textMeasurer, ctx.Path[len(ctx.Path)-1], pos, callback)
}

func (t *Touch) Position(m TextMeasurer, w interface{}) vec2.Vec2 {
	return ToLocal(m, []interface{}{w}, t.Curr)
}

func (t *Touch) StartAction(queryCtx TraversalContext, root interface{}, keycode string, key string) {
	if _, ok := t.actions[keycode]; ok {
		return
	}
	opts := QueryOptions(queryCtx, root, t.Curr)
	path, opt := PickOption(opts, keycode, key)
	if path == nil || opt == nil {
		return
	}
	ctx := TouchContext{queryCtx, t, path}
	act := opt.Activate(ctx)
	if act == nil {
		return
	}
	t.actions[keycode] = actionContext{path, act}
}

func (t *Touch) EndAction(m TextMeasurer, keycode string) {
	if entry, ok := t.actions[keycode]; ok {
		entry.action.End(TouchContext{m, t, entry.path})
		delete(t.actions, keycode)
	}
}

func (t *Touch) Move(m TextMeasurer, curr vec2.Vec2) {
	t.Last = t.Curr
	t.Curr = curr
	for _, entry := range t.actions {
		preMoveAction, ok := entry.action.(PreMoveAction)
		if !ok {
			continue
		}
		preMoveAction.PreMove(TouchContext{m, t, entry.path})
	}
	for keycode, entry := range t.actions {
		newAct := entry.action.Move(TouchContext{m, t, entry.path})
		if newAct == nil {
			delete(t.actions, keycode)
		} else if newAct != entry.action {
			t.actions[keycode] = actionContext{entry.path, newAct}
		}
	}
}
