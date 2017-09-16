package mvm

// On-screen element that can be deleted
type Deletable interface {
	Delete(Touch) Touching
}

func (c *Container) Delete(touch Touch) Touching {
	for _, elem := range c.elements {
		if del, ok := elem.(Deletable); ok {
			if t := del.Delete(touch); t != nil {
				return t
			}
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
