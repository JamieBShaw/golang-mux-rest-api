package data

import "testing"

func TestChecksValidation(t *testing.T) {
	p := &Product{

		Name:  "Jamie",
		Price: 1.00,
		SKU:   "abs-abc-def",
	}

	err := p.Validate()

	if err != nil {
		t.Fatal(err)
	}
}