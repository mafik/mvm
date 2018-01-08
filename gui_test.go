package mvm

import (
	"math"
	"math/rand"
	"testing"

	"github.com/mafik/mvm/matrix"
	"github.com/mafik/mvm/ui"
	. "github.com/mafik/mvm/vec2"
)

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

func (tc *TestCase) Resize(width, height float64) {
	e := Event{
		Type:   "Size",
		Width:  width,
		Height: height,
		Client: &tc.fc,
	}
	ProcessEvent(e)
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
	tc.bp = MakeBlueprint("test")
	tc.bp.transform.Translate(Vec2{500, 500})
	tc.fc = FakeClient{
		ev: make(chan Event, 1000),
		up: make(chan string, 1000),
	}
	tc.Resize(1000, 1000)
	TheVM.root = MakeObject(tc.bp, nil, nil)
	tc.bp.Instantiate(TheVM.root)
	return
}

func textEditSequence(tc *TestCase) {
	tc.Type("KeyF", "f")
	tc.Type("Tab", "")
	tc.Type("KeyA", "a")
	tc.Type("KeyB", "b")
	tc.Type("Backspace", "Backspace")
	tc.Type("Enter", "Enter")
	tc.Type("KeyC", "c")
	tc.Type("Tab", "")
	tc.Type("KeyF", "f")
}

func TestMatrix(t *testing.T) {
	type TestCase struct {
		input     Vec2
		transform matrix.Matrix
		output    Vec2
	}
	cases := []TestCase{
		TestCase{Vec2{2, 2}, matrix.Identity(), Vec2{2, 2}},
		TestCase{Vec2{2, 2}, matrix.Scale(2), Vec2{4, 4}},
		TestCase{Vec2{2, 2}, matrix.Translate(Vec2{-5, 1}), Vec2{-3, 3}},
		TestCase{Vec2{2, 2}, matrix.Multiply(matrix.Translate(Vec2{-5, 1}), matrix.Scale(2)), Vec2{-6, 6}},
		TestCase{Vec2{2, 2}, matrix.Multiply(matrix.Scale(2), matrix.Translate(Vec2{-5, 1})), Vec2{-1, 5}},
	}
	for i, testCase := range cases {
		result := matrix.Apply(testCase.transform, testCase.input)
		diff := Len(Sub(result, testCase.output))
		if diff > 0 {
			t.Error("Test case", i, "Expected", testCase.output, "Got", result)
		}
	}
	for i := 0; i < 10; i++ {
		r := func() float64 {
			return rand.Float64()*100 - 50
		}
		m := matrix.Matrix{r(), r(), r(), r(), r(), r()}
		inv := matrix.Invert(m)
		x := matrix.Multiply(m, inv)
		eq := func(a, b float64) bool {
			return math.Abs(a-b) < 0.000001
		}
		if !eq(x[0], 1) || !eq(x[1], 0) || !eq(x[2], 0) || !eq(x[3], 1) || !eq(x[4], 0) || !eq(x[5], 0) {
			t.Error("Matrix inverse was wrong! x =", x)
		}
	}
}

func TestTypes(t *testing.T) {
	ok := false
	var parent ui.Parent = nil
	_, ok = parent.(*ClientUI)
	_, ok = parent.(BlueprintWidget)
	if ok {
	}
}

func TestFrameRename(t *testing.T) {
	tc := setupTest()
	f := tc.bp.AddFrame()
	if f.name != "" {
		t.Error("Initial frame name is not empty")
	}
	f.pos = Vec2{50, 50}
	f.size = Vec2{100, 100}
	tc.PointAt(10, -10)
	textEditSequence(&tc)
	if f.name != "ac" {
		t.Error("Frame name is:", f.name, ", instead of \"ac\"")
	}
}

func TestPan(t *testing.T) {
	tc := setupTest()
	tc.PointAt(20, 20)
	a := tc.bp.transform[5]
	ProcessEvent(Event{
		Type:   "KeyDown",
		Code:   "KeyD",
		Key:    "D",
		Client: &tc.fc,
	})
	tc.PointAt(20, 0)
	ProcessEvent(Event{
		Type:   "KeyUp",
		Code:   "KeyD",
		Key:    "D",
		Client: &tc.fc,
	})
	b := tc.bp.transform[5]
	if a-b != -20 {
		t.Fail()
	}
}

func TestBlueprintRename(t *testing.T) {
	tc := setupTest()
	if tc.bp.name != "test" {
		t.Error("Initial Blueprint name should be \"test\"")
	}
	tc.PointAt(20, 20)
	textEditSequence(&tc)
	if tc.bp.name != "testac" {
		t.Error("\"testac\" !=", tc.bp.name)
	}
}

/*
func TestBreadcrumbIgnoresEditKeys(t *testing.T) {
	tc := setupTest()
	for code, _ := range editKeys {
		e := Event{
			Code:   code,
			Client: &tc.fc,
		}
		if BreadcrumbInput(nil, &Pointer, e) != nil {
			t.Error("Breadcrumb responded to code", code)
		}
	}
}
*/
