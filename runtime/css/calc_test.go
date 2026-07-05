package css

import "testing"

// A12: calc()/min()/max()/clamp() over cells and %.

func evalOK(t *testing.T, expr string, base int) int {
	t.Helper()
	v, ok := EvalCalc(expr, base)
	if !ok {
		t.Fatalf("EvalCalc(%q, %d) failed", expr, base)
	}
	return v
}

func TestCalcArithmetic(t *testing.T) {
	cases := []struct {
		expr string
		want int
	}{
		{"calc(10 + 5)", 15},
		{"calc(10 - 4)", 6},
		{"calc(3 * 4)", 12},
		{"calc(10 / 2)", 5},
		{"calc(2 + 3 * 4)", 14},
		{"calc((2 + 3) * 4)", 20},
		{"calc(10cell + 5ch)", 15},
	}
	for _, c := range cases {
		if got := evalOK(t, c.expr, -1); got != c.want {
			t.Errorf("%s = %d, want %d", c.expr, got, c.want)
		}
	}
}

func TestCalcPercentAgainstBase(t *testing.T) {
	if got := evalOK(t, "calc(100% - 10)", 80); got != 70 {
		t.Errorf("calc(100%% - 10) of 80 = %d, want 70", got)
	}
	if got := evalOK(t, "calc(50% + 2)", 40); got != 22 {
		t.Errorf("got %d, want 22", got)
	}
}

func TestCalcPercentWithoutBaseFails(t *testing.T) {
	if _, ok := EvalCalc("calc(50% - 1)", -1); ok {
		t.Error("percent without base must fail")
	}
}

func TestMinMaxClamp(t *testing.T) {
	cases := []struct {
		expr string
		base int
		want int
	}{
		{"min(10, 20)", -1, 10},
		{"max(10, 20, 15)", -1, 20},
		{"clamp(10, 5, 30)", -1, 10},
		{"clamp(10, 50, 30)", -1, 30},
		{"clamp(10, 20, 30)", -1, 20},
		{"min(50%, 30)", 80, 30},
		{"max(50%, 30)", 80, 40},
	}
	for _, c := range cases {
		if got := evalOK(t, c.expr, c.base); got != c.want {
			t.Errorf("%s = %d, want %d", c.expr, got, c.want)
		}
	}
}

func TestNestedCalcFunctions(t *testing.T) {
	if got := evalOK(t, "calc(min(10, 20) + max(1, 2))", -1); got != 12 {
		t.Errorf("got %d, want 12", got)
	}
}

func TestCalcInvalid(t *testing.T) {
	for _, expr := range []string{"calc(", "calc(1 +)", "calc(x)", "blah(1)", "calc(1 / 0)"} {
		if _, ok := EvalCalc(expr, -1); ok {
			t.Errorf("%q should fail", expr)
		}
	}
}

func TestIsCalcValue(t *testing.T) {
	for _, v := range []string{"calc(1 + 2)", "min(1, 2)", "max(1, 2)", "clamp(1, 2, 3)"} {
		if !IsCalcValue(v) {
			t.Errorf("%q should be a calc value", v)
		}
	}
	if IsCalcValue("20cell") || IsCalcValue("var(--x)") {
		t.Error("non-calc values misdetected")
	}
}
