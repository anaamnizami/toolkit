package toolkit

import "crypto/rand"

const SourceString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!_$123456789"

type Toolkit struct{}

func (t *Toolkit) RandomString(num int) string {
	s, r := make([]rune, num), []rune(SourceString)
	for i := range s {
		n, _ := rand.Prime(rand.Reader, len(r))
		x, y := n.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}
