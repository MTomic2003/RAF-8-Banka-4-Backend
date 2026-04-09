package service

import "math"

// riskFreeRate is the annualized risk-free interest rate used for pricing.
const riskFreeRate = 0.05

// BlackScholesCall returns the theoretical price of a European call option.
//
//	S     = current stock price
//	K     = strike price
//	T     = time to expiration in years
//	r     = risk-free interest rate
//	sigma = volatility of the underlying

func BlackScholesCall(S, K, T, r, sigma float64) float64 {
	if K <= 0 || S <= 0 {
		return math.Max(S-K, 0)
	}
	if T <= 0 || sigma <= 0 {
		return math.Max(S-K, 0)
	}
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)
	return S*cdf(d1) - K*math.Exp(-r*T)*cdf(d2)
}

func BlackScholesPut(S, K, T, r, sigma float64) float64 {
	if K <= 0 || S <= 0 {
		return math.Max(K-S, 0)
	}
	if T <= 0 || sigma <= 0 {
		return math.Max(K-S, 0)
	}
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)
	return K*math.Exp(-r*T)*cdf(-d2) - S*cdf(-d1)
}

// cdf approximates the standard normal cumulative distribution function.
func cdf(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}
