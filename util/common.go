package util

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sudachen/pprof/profile"
	"os"
)

const defaultSource = ""

type SortBy byte
const (
	ByFlat SortBy = iota
	ByCum
)

type Selector byte
const (
	Default Selector = iota
	Exclude
	Only
)

type Unit byte
const (
	Second 		Unit = 0
	Millisecond Unit = 1
	Microsecond Unit = 2
)

var defaultUnit = Second

func UnitToString(u Unit) string {
	switch u {
	case Microsecond: return "us"
	case Millisecond: return "ms"
	case Second: return "s"
	default: return UnitToString(defaultUnit)
	}
}

func UnitFromString(u *Unit, s string) error {
	switch s {
	case "us": *u = Microsecond; return nil
	case "ms": *u = Millisecond; return nil
	case "s":  *u = Second; return nil
	}
	return errors.New("invalid unit string")
}

type Options struct {
	Unit
	Sort     SortBy
	Runtime  Selector
	TagFocus []string
	Focus    []string
	Show     []string
	Hide     []string
}

type Row struct {
	Flat, FlatPercent, SumPercent, Cum, CumPercent float64
	Function                                       string
}

type Rows []*Row

type Report struct {
	Unit
	Rows
	Label  string
	Errors []string
	Image  string
}

func tuneBy(o *Options) []string {
	var c []string

	show := make([]string,0,3)
	hide := make([]string,0,3)

	c = append(c, "unit="+UnitToString(o.Unit))

	if o.Sort == ByCum {
		c = append(c, "cum=true")
	} else {
		c = append(c, "flat=true")
	}

	switch o.Runtime {
	case Default:
	case Exclude:
		hide = append(hide,"^runtime\\..*$")
	case Only:
		show = append(show,"^runtime\\..*$")
	}

	if len(o.Hide) != 0 {
		hide = append(hide,o.Hide...)
	}
	c = append(c, "hide="+strings.Join(hide,"|"))

	if len(o.Show) != 0 {
		hide = append(hide,o.Show...)
	}
	c = append(c, "show="+strings.Join(show,"|"))

	if len(o.TagFocus) != 0 {
		s := "tagfocus="+o.TagFocus[0]
		for _, t := range(o.TagFocus[1:]) {
			s = s + "|" + t
		}
		c = append(c, s)
	} else {
		c = append(c, "tagfocus=")
	}

	if len(o.Focus) != 0 {
		s := "focus="+o.Focus[0]
		for _, t := range(o.Focus[1:]) {
			s = s + "|" + t
		}
		c = append(c, s)
	} else {
		c = append(c, "focus=")
	}

	fmt.Fprintln(os.Stderr,c)

	return c
}

type fetcher struct {
	b []byte
}

func (f *fetcher) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	if src == defaultSource {
		p, err := profile.ParseData(f.b)
		return p, "", err
	}
	return nil, "", fmt.Errorf("unknown source %s", src)
}

type ui struct {
	*Report
	command []string
	index   int
}

func (u *ui) ReadLine(prompt string) (string, error) {
	if u.index < len(u.command) {
		u.index++
		return u.command[u.index-1], nil
	}
	return "quit", nil
}

func (u *ui) PrintErr(a ...interface{}) {
	if u.Report != nil {
		u.Report.Errors = append(u.Report.Errors, fmt.Sprint(a...))
	}
}

func (u *ui) Print(a ...interface{})                       {}
func (u *ui) IsTerminal() bool                             { return false }
func (u *ui) SetAutoComplete(complete func(string) string) {}

