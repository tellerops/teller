package crypto

// InsecureStream is a fake cipher that implements CryptoStream interface
type InsecureStream struct{}

// NewInsecureStream initialize a new InsecureStream interfaced with CryptoStream
func NewInsecureStream() *InsecureStream {
	return new(InsecureStream)
}

// Unpack returns the payload as unencrypted byte array
func (c *InsecureStream) Unpack(payload string) []byte {
	return []byte(payload)
}

// Pack returns the payload as encrypted string
func (c *InsecureStream) Pack(payload []byte) string {
	return string(payload)
}
