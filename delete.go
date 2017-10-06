package mvm

// On-screen element that can be deleted
type Deletable interface {
	Delete(Touch) Touching
}

func (c *LayerList) Delete(touch Touch) Touching {
	for _, elem := range *c {
		if t := elem.Delete(touch); t != nil {
			return t
		}
	}
	return nil
}

func (FrameLayer) Delete(t Touch) Touching {
	deleted := t.FindFrameBelow()
	if deleted == nil {
		return nil
	}
	deleted.Delete()
	return NoopTouching{}
}

func (LinkLayer) Delete(t Touch) Touching {
	deleted := t.PointedLink()
	if deleted == nil {
		return nil
	}
	deleted.Delete()
	return NoopTouching{}
}

func (ParamLayer) Delete(t Touch) Touching {
	frame, name := Pointer.FindParamBelow()
	if frame == nil {
		return nil
	}
	index, _ := GetParam(frame.LocalParameters(), name)
	if index == -1 {
		return nil
	}
	frame.params = append(frame.params[:index], frame.params[index+1:]...)
	return NoopTouching{}
}

func (ObjectLayer) Delete(t Touch) Touching     { return nil }
func (OverlayLayer) Delete(t Touch) Touching    { return nil }
func (ParamNameLayer) Delete(t Touch) Touching  { return nil }
func (BackgroundLayer) Delete(t Touch) Touching { return nil }
