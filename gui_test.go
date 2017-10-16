package mvm

import "testing"

type TestCase struct {
	bp *Blueprint
	fc FakeClient
}

type FakeClient struct {
	ev chan Event
	up chan string
}

func (c *FakeClient) Call(req string) (resp Event, err error) {
	return
}

func (tc *TestCase) PointAt(x, y float64) {
	e := Event{
		Type:   "MouseMove",
		X:      x,
		Y:      y,
		Client: &tc.fc,
	}
	ProcessEvent(e)
}

func (tc *TestCase) Type(code, key string) {
	ProcessEvent(Event{
		Type:   "KeyDown",
		Code:   code,
		Key:    key,
		Client: &tc.fc,
	})
	ProcessEvent(Event{
		Type:   "KeyUp",
		Code:   code,
		Key:    key,
		Client: &tc.fc,
	})
}

func setupTest() (tc TestCase) {
	TheVM = &VM{}
	window = Window{1, Vec2{1000, 1000}, Vec2{500, 500}}
	tc.bp = MakeBlueprint("test")
	tc.fc = FakeClient{
		ev: make(chan Event, 1000),
		up: make(chan string, 1000),
	}
	TheVM.active = MakeObject(tc.bp, nil, nil)
	tc.bp.Instantiate(TheVM.active)
	return
}

func textEditSequence(tc *TestCase) {
	tc.Type("KeyA", "a")
	tc.Type("KeyB", "b")
	tc.Type("Backspace", "Backspace")
	tc.Type("Enter", "Enter")
	tc.Type("KeyC", "c")
}

func TestFrameRename(t *testing.T) {
	tc := setupTest()
	f := tc.bp.AddFrame()
	if f.name != "" {
		t.Fail()
	}
	f.pos = Vec2{50, 50}
	f.size = Vec2{100, 100}
	tc.PointAt(10, -10)
	textEditSequence(&tc)
	if f.name != "ac" {
		t.Fail()
	}
}

func TestBlueprintRename(t *testing.T) {
	tc := setupTest()
	if tc.bp.name != "test" {
		t.Fail()
	}
	tc.PointAt(20, 20)
	textEditSequence(&tc)
	if tc.bp.name != "testac" {
		t.Fail()
	}
}
