package util

import (
	"fmt"
	"flag"
	"github.com/sudachen/pprof/driver"
	"io"
	"bytes"
)

type pngWriter struct{ *bytes.Buffer }

func (w *pngWriter) Open(name string) (io.WriteCloser, error) {
	return w, nil
}

func (w *pngWriter) Close() error {
	return nil
}

func Png(b []byte, count int, o *Options) []byte {
	var bf bytes.Buffer
	unit := o.Unit
	if unit == 0 { unit = defaultUnit }
	c := append(tuneBy(o), "output=@", fmt.Sprintf("nodecount=%d",count), "png")
	f := &FlagSet{flag.NewFlagSet("ppf", flag.ContinueOnError), []string{defaultSource}}
	d := &driver.Options{
		Flagset: f,
		Fetch:   &fetcher{b},
		Writer:  &pngWriter{&bf},
		UI:      &ui{nil, c, 0},
	}
	driver.PProf(d)
	return bf.Bytes()
}

