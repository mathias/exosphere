package docker

import (
	"bufio"
	"io"
	"strings"

	"github.com/Originate/exosphere/src/util"
)

// Takes an io.ReadCloser and returns a channel that closes when the stream ends
// Used to determine when an image is finished pulling
func readCloserToChan(readCloser io.ReadCloser) chan int {
	done := make(chan int)
	go func() {
		b := make([]byte, 256)
		var err error
		for err == nil {
			_, err = readCloser.Read(b)
		}
		close(done)
	}()
	return done
}

func logReader(reader io.Reader, logger *util.Logger) {
	bufioReader := bufio.NewReader(reader)
	for {
		line, _, err := bufioReader.ReadLine()
		if err != nil {
			break
		}
		logger.LogNew(string(line))
	}
}

func formateImageName(name string) string {
	return strings.ToLower(name)
}
