package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipeSuccess(t *testing.T) {

	for i := 0; i < 1000; i++ {
		func() {
			e := NewPipe(func(w io.Writer) error {
				return json.NewEncoder(w).Encode(map[string]interface{}{
					"k1": "val1",
					"k2": "val2",
				})
			})
			defer func() { require.NoError(t, e.Close()) }()

			buf := bytes.NewBuffer(nil)
			chunk := make([]byte, 1000)
			for {
				n, err := e.Read(chunk)
				if err == io.EOF {
					break
				} else if err != nil {
					require.Fail(t, err.Error())
				}

				buf.Write(chunk[:n])
			}

			require.Equal(t, `{"k1":"val1","k2":"val2"}`+"\n", buf.String())

			n, err := e.Read(chunk)
			require.Equal(t, 0, n)
			require.Equal(t, io.EOF, err)
		}()
	}
}

func TestPipeReadAndClose(t *testing.T) {

	for i := 0; i < 1000; i++ {
		func() {
			e := NewPipe(func(w io.Writer) error {
				return json.NewEncoder(w).Encode(map[string]interface{}{
					"k1": "val1",
					"k2": "val2",
				})
			})
			defer func() { require.NoError(t, e.Close()) }()

			buf := bytes.NewBuffer(nil)
			chunk := make([]byte, 2)
			for {
				n, err := e.Read(chunk)
				if err == io.EOF {
					break
				} else if err != nil {
					require.Fail(t, err.Error())
				}

				buf.Write(chunk[:n])
				require.NoError(t, e.Close())
			}

			require.Equal(t, `{"`, buf.String())

			n, err := e.Read(chunk)
			require.Equal(t, 0, n)
			require.Equal(t, io.EOF, err)
		}()
	}
}
