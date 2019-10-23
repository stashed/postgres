package restic

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"stash.appscode.dev/stash/apis/stash/v1alpha1"

	shell "github.com/codeskyblue/go-sh"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	DefaultOutputFileName = "output.json"
	DefaultScratchDir     = "/tmp"
	DefaultHost           = "host-0"
)

type ResticWrapper struct {
	sh     *shell.Session
	config SetupOptions
}

type Command struct {
	Name string
	Args []interface{}
}

// BackupOptions specifies backup information
// if StdinPipeCommand is specified, BackupPaths will not be used
type BackupOptions struct {
	Host             string
	BackupPaths      []string
	StdinPipeCommand Command
	StdinFileName    string // default "stdin"
	RetentionPolicy  v1alpha1.RetentionPolicy
}

// RestoreOptions specifies restore information
type RestoreOptions struct {
	Host         string
	SourceHost   string
	RestorePaths []string
	Snapshots    []string // when Snapshots are specified SourceHost and RestorePaths will not be used
	Destination  string   // destination path where snapshot will be restored, used in cli
}

type DumpOptions struct {
	Host              string
	SourceHost        string
	Snapshot          string // default "latest"
	Path              string
	FileName          string // default "stdin"
	StdoutPipeCommand Command
}

type SetupOptions struct {
	Provider       string
	Bucket         string
	Endpoint       string
	Path           string
	SecretDir      string
	CacertFile     string
	ScratchDir     string
	EnableCache    bool
	MaxConnections int
	Nice           *ofst.NiceSettings
	IONice         *ofst.IONiceSettings
}

type MetricsOptions struct {
	Enabled        bool
	PushgatewayURL string
	MetricFileDir  string
	Labels         []string
	JobName        string
}

func NewResticWrapper(options SetupOptions) (*ResticWrapper, error) {
	wrapper := &ResticWrapper{
		sh:     shell.NewSession(),
		config: options,
	}
	wrapper.sh.SetDir(wrapper.config.ScratchDir)
	wrapper.sh.ShowCMD = true
	wrapper.sh.PipeFail = true
	wrapper.sh.PipeStdErrors = true

	// Setup restic environments
	err := wrapper.setupEnv()
	if err != nil {
		return nil, err
	}
	return wrapper, nil
}

func (w *ResticWrapper) SetEnv(key, value string) {
	if w.sh != nil {
		w.sh.SetEnv(key, value)
	}
}

func (w *ResticWrapper) DumpEnv(path string, dumpedFile string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	var envs string
	if w.sh != nil {
		sortedKeys := make([]string, 0, len(w.sh.Env))
		for k := range w.sh.Env {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys) //sort by key
		for _, v := range sortedKeys {
			envs = envs + fmt.Sprintln(v+"="+w.sh.Env[v])
		}
	}

	if err := ioutil.WriteFile(filepath.Join(path, dumpedFile), []byte(envs), 0600); err != nil {
		return err
	}
	return nil
}

func (w *ResticWrapper) HideCMD() {
	if w.sh != nil {
		w.sh.ShowCMD = false
	}
}

func (w *ResticWrapper) GetRepo() string {
	if w.sh != nil {
		return w.sh.Env[RESTIC_REPOSITORY]
	}
	return ""
}

// Copy function copy input ResticWrapper and returns a new wrapper with copy of its content.
func (w *ResticWrapper) Copy() *ResticWrapper {
	if w == nil {
		return nil
	}
	out := new(ResticWrapper)

	if w.sh != nil {
		out.sh = shell.NewSession()

		// set values in.sh to out.sh
		for k, v := range w.sh.Env {
			out.sh.Env[k] = v
		}
		// don't use same stdin, stdout, stderr for each instant to avoid data race.
		//out.sh.Stdin = in.sh.Stdin
		//out.sh.Stdout = in.sh.Stdout
		//out.sh.Stderr = in.sh.Stderr
		out.sh.ShowCMD = w.sh.ShowCMD
		out.sh.PipeFail = w.sh.PipeFail
		out.sh.PipeStdErrors = w.sh.PipeStdErrors

	}
	out.config = w.config
	return out
}
