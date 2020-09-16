package transaction

type Output struct {
	Value         int
	PublicKeyHash []byte
}

type Outputs []Output

func (outs Outputs) Find(criteria func(Output) bool) (Output, bool) {
	for _, out := range outs {
		if criteria(out) {
			return out, true
		}
	}
	return Output{}, false
}

func (outs Outputs) Sum() (sum int) {
	for _, out := range outs {
		sum += out.Value
	}
	return sum
}
