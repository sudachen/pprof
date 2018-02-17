package util

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/sudachen/pprof/util/driver"
	"github.com/sudachen/pprof/internal/measurement"
	"github.com/sudachen/pprof/profile"
)

func isLocalhost(host string) bool {
	for _, v := range []string{"localhost", "127.0.0.1", "[::1]", "::1"} {
		if host == v {
			return true
		}
	}
	return false
}

type currentProfile struct {
	bf   bytes.Buffer
	prof *profile.Profile
	mu   sync.Mutex
	stop chan interface{}
	serv *http.Server
}

func (f *currentProfile) Stop() {
	close(f.stop)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	f.serv.Shutdown(ctx)
}

func (f *currentProfile) Serve(args *driver.HTTPServerArgs) error {
	ln, err := net.Listen("tcp", args.Hostport)
	if err != nil {
		return err
	}
	isLocal := isLocalhost(args.Host)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if isLocal {
			// Only allow local clients
			host, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil || !isLocalhost(host) {
				http.Error(w, "permission denied", http.StatusForbidden)
				return
			}
		}
		h := args.Handlers[req.URL.Path]
		if h == nil {
			// Fall back to default behavior
			h = http.DefaultServeMux
		}
		h.ServeHTTP(w, req)
	})
	f.serv = &http.Server{Handler: handler}
	return f.serv.Serve(ln)
}

var emptyProfile *profile.Profile
var globalProfile *currentProfile

func (f *currentProfile) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	switch src {
	case "-": // hack for web UI. Restart profile updating from scratch
		f.mu.Lock()
		f.prof = emptyProfile
		f.mu.Unlock()
		src = defaultSource
		fallthrough
	case defaultSource:
		f.mu.Lock()
		prof := f.prof
		f.mu.Unlock()
		if prof == nil {
			prof = emptyProfile
		}
		return prof, "", nil
	default:
		return nil, "", fmt.Errorf("unknown source %s", src)
	}
}

func (f *currentProfile) Update(b []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p, err := profile.ParseData(b); err != nil {
		return err
	} else {
		if f.prof == nil {
			f.prof = p
		} else {
			pfs := []*profile.Profile{f.prof, p}
			if err := measurement.ScaleProfiles(pfs); err != nil {
				return err
			}
			if p, err := profile.Merge(pfs); err != nil {
				return err
			} else {
				p.RemoveUninteresting()
				if err := p.CheckValid(); err != nil {
					return err
				}
				f.prof = p
			}
		}
	}
	return nil
}

func (f *currentProfile) UpdateLoop(interval time.Duration) {
	pprof.StartCPUProfile(&f.bf)
	for {
		select {
		case <-f.stop:
			return
		case <-time.After(interval):
			pprof.StopCPUProfile()
			if err := f.Update(f.bf.Bytes()); err != nil {
				fmt.Fprintf(os.Stderr, "failed to merge profiles: %s", err.Error())
			}
			f.bf.Reset()
			pprof.StartCPUProfile(&f.bf)
		}
	}
}

func (*currentProfile) ReadLine(prompt string) (string, error) {
	return "quit", nil
}

func (*currentProfile) PrintErr(a ...interface{})                    {}
func (*currentProfile) Print(a ...interface{})                       {}
func (*currentProfile) IsTerminal() bool                             { return false }
func (*currentProfile) SetAutoComplete(complete func(string) string) {}

func Start(interval time.Duration, port int) {
	if globalProfile == nil {
		globalProfile := &currentProfile{stop: make(chan interface{})}

		if emptyProfile == nil {
			pprof.StartCPUProfile(&globalProfile.bf)
			pprof.StopCPUProfile()
			emptyProfile, _ = profile.ParseData(globalProfile.bf.Bytes())
		}

		go globalProfile.UpdateLoop(interval)
		go driver.PProf(&driver.Options{
			Fetch: globalProfile,
			UI:    globalProfile,
			HTTPServer: func(args *driver.HTTPServerArgs) error {
				return globalProfile.Serve(args)
			},
			Flagset: &FlagSet{
				FlagSet: flag.NewFlagSet("rtppf", flag.ContinueOnError),
				Args:    []string{fmt.Sprintf("--http=:%d", port), defaultSource},
			},
		})
	}
}

func Stop() {
	if globalProfile != nil {
		globalProfile.Stop()
		globalProfile = nil
	}
}
