package ui

type Parent interface {
	Children() []interface{} // stable back-to-front order
}

type WidgetPath []interface{} // root-to-child

type WalkAction int

const (
	Explore WalkAction = iota
	Avoid
	Return
)

type EnterCallback func(WidgetPath) WalkAction
type ExitCallback func(WidgetPath)

func NoopExit(WidgetPath) {}

func WalkWithin(path WidgetPath, enter EnterCallback, exit ExitCallback) WalkAction {
	if exit == nil {
		exit = NoopExit
	}
	switch enter(path) {
	case Explore:
		if parent, ok := path[len(path)-1].(Parent); ok {
			for _, child := range parent.Children() {
				childPath := append(path, child)
				if WalkWithin(childPath, enter, exit) == Return {
					exit(path)
					return Return
				}
			}
		}
	case Return:
		exit(path)
		return Return
	}
	exit(path)
	return Explore
}

func Walk(start interface{}, enter EnterCallback, exit ExitCallback) {
	stack := []interface{}{start}
	WalkWithin(stack, enter, exit)
}
