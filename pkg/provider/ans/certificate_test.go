package ans

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTopics(t *testing.T) {

	{
		topics, err := GetTopics([]byte{
			0x30, 0x18, 0xc, 0x1, 'a',
			0x30, 0x2, 0xc, 0x2, 'b', 'c',
			0xc, 0x01, 'd',
			0x30, 0x3, 0xc, 0x3, 'e', 'f', 'g',
			0xc, 0x03, 'h', 'i', 'j'})
		require.NoError(t, err)
		require.Equal(t, []string{"a", "bc", "d", "efg", "hij"}, topics)
	}

	{
		topics, err := GetTopics([]byte{
			0x30, 0x3, 0xc, 0x1, 'a'})
		require.NoError(t, err)
		require.Equal(t, []string{"a"}, topics)
	}

	{
		topics, err := GetTopics([]byte{
			0x30, 0x8, 0xc, 0x1, 'a',
			0xc, 0x03, 'h', 'i', 'j'})
		require.NoError(t, err)
		require.Equal(t, []string{"a", "hij"}, topics)
	}

}
