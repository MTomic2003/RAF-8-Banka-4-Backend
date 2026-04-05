package service

import (
	"math"
	"testing"
)

func TestBlackScholesCallPut(t *testing.T) {
	// Classic textbook example: S=100, K=100, T=1yr, r=5%, σ=20%
	call := BlackScholesCall(100, 100, 1.0, 0.05, 0.20)
	put := BlackScholesPut(100, 100, 1.0, 0.05, 0.20)

	// Expected values (well-known): call ≈ 10.45, put ≈ 5.57
	if math.Abs(call-10.45) > 0.1 {
		t.Errorf("call = %.4f, want ≈ 10.45", call)
	}
	if math.Abs(put-5.57) > 0.1 {
		t.Errorf("put = %.4f, want ≈ 5.57", put)
	}

	// Put-call parity: C - P = S - K*e^(-rT)
	parity := call - put - (100 - 100*math.Exp(-0.05))
	if math.Abs(parity) > 0.001 {
		t.Errorf("put-call parity violated: diff = %.6f", parity)
	}
}

func TestBlackScholesExpired(t *testing.T) {
	// T=0 → intrinsic value
	call := BlackScholesCall(105, 100, 0, 0.05, 0.20)
	if math.Abs(call-5.0) > 0.001 {
		t.Errorf("expired ITM call = %.4f, want 5.0", call)
	}

	put := BlackScholesPut(95, 100, 0, 0.05, 0.20)
	if math.Abs(put-5.0) > 0.001 {
		t.Errorf("expired ITM put = %.4f, want 5.0", put)
	}
}
