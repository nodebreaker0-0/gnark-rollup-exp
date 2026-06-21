// Package cubic is the zkkit "hello world": prove knowledge of a secret x such
// that x^3 + x + 5 == y, revealing only the public output y.
package cubic

import "github.com/consensys/gnark/frontend"

// Circuit proves x^3 + x + 5 == y. X is a secret input; Y is public.
type Circuit struct {
	X frontend.Variable `gnark:"x"`        // secret witness
	Y frontend.Variable `gnark:"y,public"` // public output
}

// Define declares the constraint x^3 + x + 5 == y.
func (c *Circuit) Define(api frontend.API) error {
	x3 := api.Mul(c.X, c.X, c.X)
	api.AssertIsEqual(c.Y, api.Add(x3, c.X, 5))
	return nil
}
