package crypto

import "testing"

// key hex ขนาด 32 ไบต์สำหรับทดสอบ
const testKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func newTestCipher(t *testing.T) *Cipher {
	t.Helper()
	c, err := NewCipher(testKey)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	return c
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	c := newTestCipher(t)
	plain := "1234567890123"

	ct, err := c.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if string(ct) == plain {
		t.Fatal("ciphertext ต้องไม่เท่ากับ plaintext")
	}

	got, err := c.Decrypt(ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("decrypt ได้ %q ต้องการ %q", got, plain)
	}
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	c := newTestCipher(t)
	a, _ := c.Encrypt("1234567890123")
	b, _ := c.Encrypt("1234567890123")
	if string(a) == string(b) {
		t.Fatal("เข้ารหัสค่าเดียวกันสองครั้งต้องได้ ciphertext ต่างกัน (nonce สุ่ม)")
	}
}

func TestHashIsDeterministic(t *testing.T) {
	c := newTestCipher(t)
	if c.Hash("1234567890123") != c.Hash("1234567890123") {
		t.Fatal("Hash ค่าเดียวกันต้องได้ผลเท่ากัน (ใช้ทำ unique/ค้นหา)")
	}
	if c.Hash("1234567890123") == c.Hash("9999999999999") {
		t.Fatal("Hash ค่าต่างกันต้องได้ผลต่างกัน")
	}
}

func TestNewCipherRejectsBadKey(t *testing.T) {
	if _, err := NewCipher("tooshort"); err == nil {
		t.Fatal("key สั้นต้อง error")
	}
}

func TestMask(t *testing.T) {
	got := Mask("1234567890123")
	want := "1-2345-xxxxx-xx-3"
	if got != want {
		t.Fatalf("Mask ได้ %q ต้องการ %q", got, want)
	}
	if Mask("123") != "xxxxxxxxxxxxx" {
		t.Fatal("เลขไม่ครบ 13 หลักต้อง mask ทั้งหมด")
	}
}
