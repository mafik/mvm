package mvm

type Object struct {
	parent  *Object
	frame   *Frame
	execute bool
	running bool
	typ     Type
	priv    interface{}
}

func MakeObject(typ Type, frame *Frame, parent *Object) *Object {
	o := &Object{
		typ:    typ,
		frame:  frame,
		parent: parent,
	}
	if parent != nil {
		m := parent.priv.(*Machine)
		m.objects[frame] = o
	}
	return o
}

func (o *Object) MarkForExecution() {
	o.execute = true
	tasks <- o
}
