package base62
import (
 "testing"

 "github.com/stretchr/testify/assert"
)

func TestBase62Encode_Consistency(t *testing.T) {
 t.Run("Encode same number multiple times", func(t *testing.T) {
  number := uint64(1234567890)
  expected := Base62Encode(number)

  for i := 0; i < 10; i++ {
   actual := Base62Encode(number)
   assert.Equal(t, expected, actual, "Encoded value should be consistent")
  }
 })
}
