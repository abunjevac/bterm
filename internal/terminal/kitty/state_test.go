package kitty

import "testing"

func TestStateDefaultsToZero(t *testing.T) {
	var s State

	if s.Flags() != 0 {
		t.Fatalf("expected flags=0, got %d", s.Flags())
	}

	if s.Disambiguate() {
		t.Fatal("expected disambiguate to be false by default")
	}
}

func TestStateSetReplace(t *testing.T) {
	var s State

	s.Set(FlagDisambiguate|FlagReportEvents, 1)

	if s.Flags() != FlagDisambiguate|FlagReportEvents {
		t.Fatalf("expected flags=3, got %d", s.Flags())
	}
}

func TestStateSetUnion(t *testing.T) {
	s := State{flags: FlagDisambiguate}

	s.Set(FlagReportEvents, 2)

	if s.Flags() != FlagDisambiguate|FlagReportEvents {
		t.Fatalf("expected flags=3, got %d", s.Flags())
	}
}

func TestStateSetSubtract(t *testing.T) {
	s := State{flags: FlagDisambiguate | FlagReportEvents}

	s.Set(FlagDisambiguate, 3)

	if s.Flags() != FlagReportEvents {
		t.Fatalf("expected flags=2, got %d", s.Flags())
	}
}

func TestStatePushPop(t *testing.T) {
	var s State

	s.Set(FlagDisambiguate, 1)
	s.Push(FlagReportEvents)

	if s.Flags() != FlagReportEvents {
		t.Fatalf("after push, expected flags=2, got %d", s.Flags())
	}

	s.Pop(1)

	if s.Flags() != FlagDisambiguate {
		t.Fatalf("after pop, expected flags=1, got %d", s.Flags())
	}
}

func TestStatePushZeroClearsFlags(t *testing.T) {
	s := State{flags: FlagDisambiguate}

	s.Push(0)

	if s.Flags() != 0 {
		t.Fatalf("expected flags=0 after push(0), got %d", s.Flags())
	}
}

func TestStatePopEmptyStackResets(t *testing.T) {
	var s State

	s.Pop(1)

	if s.Flags() != 0 {
		t.Fatalf("expected flags=0, got %d", s.Flags())
	}
}

func TestStatePopMultiple(t *testing.T) {
	var s State

	s.Set(1, 1)
	s.Push(2)
	s.Push(4)
	s.Push(8)

	s.Pop(2)

	if s.Flags() != 2 {
		t.Fatalf("expected flags=2, got %d", s.Flags())
	}
}

func TestStatePopMoreThanStack(t *testing.T) {
	var s State

	s.Set(1, 1)
	s.Push(2)
	s.Push(4)

	s.Pop(10)

	if s.Flags() != 0 {
		t.Fatalf("expected flags=0, got %d", s.Flags())
	}
}

func TestStateQueryResponse(t *testing.T) {
	s := State{flags: FlagDisambiguate}

	if got := s.QueryResponse(); got != "\x1b[?1u" {
		t.Fatalf("expected \\x1b[?1u, got %q", got)
	}
}

func TestStateStackDepthLimit(t *testing.T) {
	var s State

	for i := 0; i < maxStackDepth+10; i++ {
		s.Push(i)
	}

	if len(s.stack) > maxStackDepth {
		t.Fatalf("stack grew beyond limit: %d", len(s.stack))
	}
}
