package ui

type Parent interface {
	Children() []interface{} // stable back-to-front order
}

type TreePath []interface{} // root-to-child

type WalkAction int

const (
	Explore WalkAction = iota
	Avoid
	Return
)

type EnterCallback func(TreePath) WalkAction
type ExitCallback func(TreePath)

func NoopExit(TreePath) {}

func WalkWithin(path TreePath, enter EnterCallback, exit ExitCallback) WalkAction {
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
