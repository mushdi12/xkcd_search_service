package closers

import (
	"io"
	"log/slog"
)

func CloseOrLog(log *slog.Logger, closers ...io.Closer)  {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			log.Error("close failed", "error", err)
		}
	}
}

func CloseOrPanic(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(err)
	}
}
