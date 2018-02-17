package util

import (
	"fmt"
	"flag"
	"github.com/sudachen/pprof/util/driver"
	"io"
	"strings"
)

type topWriter struct{ *Report }

func (w *topWriter) Open(name string) (io.WriteCloser, error) {
	return w, nil
}

func (w *topWriter) Write(p []byte) (n int, err error) {
	tf := func(s string) (r []string) {
		r = make([]string, 0, len(s))
		for _, v := range strings.Fields(s) {
			if s := strings.TrimSpace(v); len(s) > 0 {
				r = append(r, s)
			}
		}
		return
	}

	skip := true
	for _, l := range strings.Split(string(p), "\n") {
		a := tf(l)
		if skip && "flat flat% sum% cum cum%" == strings.Join(a, " ") {
			skip = false
		}
		if !skip && len(a) > 5 {
			i := &Row{}
			fmt.Sscanf(a[0], "%f", &i.Flat)
			fmt.Sscanf(a[1], "%f", &i.FlatPercent)
			fmt.Sscanf(a[2], "%f", &i.SumPercent)
			fmt.Sscanf(a[3], "%f", &i.Cum)
			fmt.Sscanf(a[4], "%f", &i.CumPercent)
			i.Function = a[5]
			w.Report.Rows = append(w.Report.Rows, i)
		}
	}
	return
}

func (w *topWriter) Close() error {
	return nil
}

func Top(b []byte, count int, o *Options, label string) *Report {
	unit := o.Unit
	if unit == 0 { unit = defaultUnit }
	rpt := &Report{Label: label, Unit: unit, Rows: make(Rows, 0, count)}
	c := append(tuneBy(o), "output=@", fmt.Sprintf("top%d", count))
	f := &FlagSet{flag.NewFlagSet("ppf", flag.ContinueOnError), []string{defaultSource}}
	d := &driver.Options{
		Flagset: f,
		Fetch:   &fetcher{b},
		Writer:  &topWriter{rpt},
		UI:      &ui{rpt, c, 0},
	}
	driver.PProf(d)
	return rpt
}

