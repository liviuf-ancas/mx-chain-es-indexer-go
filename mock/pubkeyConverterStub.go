package mock

// PubkeyConverterStub -
type PubkeyConverterStub struct {
	LenCalled         func() int
	DecodeCalled      func(humanReadable string) ([]byte, error)
	EncodeCalled      func(pkBytes []byte) (string, error)
	EncodeSliceCalled func(pkBytesSlice [][]byte) ([]string, error)
}

// Len -
func (pcs *PubkeyConverterStub) Len() int {
	if pcs.LenCalled != nil {
		return pcs.LenCalled()
	}

	return 0
}

// Decode -
func (pcs *PubkeyConverterStub) Decode(humanReadable string) ([]byte, error) {
	if pcs.DecodeCalled != nil {
		return pcs.DecodeCalled(humanReadable)
	}

	return make([]byte, 0), nil
}

// Encode -
func (pcs *PubkeyConverterStub) Encode(pkBytes []byte) (string, error) {
	if pcs.EncodeCalled != nil {
		return pcs.EncodeCalled(pkBytes)
	}

	return "", nil
}

// Encode -
func (pcs *PubkeyConverterStub) EncodeSlice(pkBytesSlice [][]byte) ([]string, error) {
	if pcs.EncodeSliceCalled != nil {
		return pcs.EncodeSliceCalled(pkBytesSlice)
	}

	return make([]string, 0), nil
}

// IsInterfaceNil -
func (pcs *PubkeyConverterStub) IsInterfaceNil() bool {
	return pcs == nil
}
