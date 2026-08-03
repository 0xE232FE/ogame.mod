package gameforge
import "testing"
func TestSolvePowKnown(t *testing.T){
    n := SolvePow("c8ea7be2079a01ba2596b138096adc264e5f1d3eff31a8658e1bf557bf77e123","00000")
    if n != 817404 { t.Fatalf("expected 817404, got %d", n) }
    n2 := SolvePowParallel("c8ea7be2079a01ba2596b138096adc264e5f1d3eff31a8658e1bf557bf77e123","00000",8)
    if n2 != 817404 { t.Fatalf("parallel expected 817404, got %d", n2) }
}
