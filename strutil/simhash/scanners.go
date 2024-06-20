package simhash

// TODO(dgryski): channel scanner
// TODO(dgryski): ngram scanner for bytes and words

// Return features one-at-a-time to be considered by SimHash.
// This matches (partially) the scanner interface for bufio.Scanner, so those scanner can be reused here.
type FeatureScanner interface {
	Scan() bool
	Bytes() []byte
	Err() error
}

type SliceScanner struct {
	tokens [][]byte
	i      int
}

// NewSliceScanner creates a scanner that returns the byte slices in tokens
func NewSliceScanner(tokens [][]byte) FeatureScanner {
	return &SliceScanner{tokens: tokens}
}

func (s *SliceScanner) Err() error {
	return nil
}

func (s *SliceScanner) Scan() bool {

	if s.i >= len(s.tokens) {
		return false
	}
	s.i++
	return true
}

func (s *SliceScanner) Bytes() []byte {
	return s.tokens[s.i-1]
}

func ScanByteTrigrams(data []byte, atEOF bool) (advance int, token []byte, err error) {

	if atEOF || len(data) < 3 {
		return 0, nil, nil
	}

	return 1, data[:3], nil
}
