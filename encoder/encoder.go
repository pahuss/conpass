package encoder

type Encoder interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	SetKey(key []byte)
}

func NewEncoder() Encoder {
	return &Aes{
		key: []byte{},
	}
}
