/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"

	shell "gomodules.xyz/go-sh"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	EnvPGSSLROOTCERT = "PGSSLROOTCERT"
	EnvPGSSLCERT     = "PGSSLCERT"
	EnvPGSSLKEY      = "PGSSLKEY"
	EnvPGSSLMODE     = "PGSSLMODE"
	EnvPgPassword    = "PGPASSWORD"
	PgDumpFile       = "dumpfile.sql"
	PgDumpCMD        = "pg_dump"
	PgDumpallCMD     = "pg_dumpall"
	PgRestoreCMD     = "psql"

	// Deprecated
	envPostgresUser = "POSTGRES_USER"
	// Deprecated
	envPostgresPassword = "POSTGRES_PASSWORD"
	DefaultPostgresUser = "postgres"
	SedCMD              = "sed"
	sedArgs             = "/ALTER ROLE postgres WITH SUPERUSER INHERIT CREATEROLE CREATEDB LOGIN REPLICATION BYPASSRLS PASSWORD/d"
)

type postgresOptions struct {
	kubeClient    kubernetes.Interface
	stashClient   stash.Interface
	catalogClient appcatalog_cs.Interface

	namespace         string
	backupSessionName string
	appBindingName    string
	backupCMD         string
	pgArgs            string
	outputDir         string
	storageSecret     kmapi.ObjectReference
	waitTimeout       int32

	setupOptions  restic.SetupOptions
	backupOptions restic.BackupOptions
	dumpOptions   restic.DumpOptions
}

func must(v []byte, err error) string {
	if err != nil {
		panic(err)
	}
	return string(v)
}

type sessionWrapper struct {
	sh  *shell.Session
	cmd *restic.Command
}

func (opt *postgresOptions) newSessionWrapper(cmd string) *sessionWrapper {
	return &sessionWrapper{
		sh: shell.NewSession(),
		cmd: &restic.Command{
			Name: cmd,
		},
	}
}

func (opt *postgresOptions) setDatabaseCredentials(appBinding *appcatalog.AppBinding, session *sessionWrapper) error {
	appBindingSecret, err := opt.kubeClient.CoreV1().Secrets(appBinding.Namespace).Get(context.TODO(), appBinding.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = appBinding.TransformSecret(opt.kubeClient, appBindingSecret.Data)
	if err != nil {
		return err
	}

	userName := ""

	if _, ok := appBindingSecret.Data[core.TLSPrivateKeyKey]; ok {
		certByte, ok := appBindingSecret.Data[core.TLSCertKey]
		if !ok {
			return fmt.Errorf("can't find client cert")
		}
		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey), certByte, 0600); err != nil {
			return err
		}

		session.sh.SetEnv(EnvPGSSLCERT, filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey))
		keyByte, ok := appBindingSecret.Data[core.TLSPrivateKeyKey]
		if !ok {
			return fmt.Errorf("can't find client private key")
		}

		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey), keyByte, 0600); err != nil {
			return err
		}
		session.sh.SetEnv(EnvPGSSLKEY, filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey))

		//TODO: this one is hard coded here but need to change later
		userName = DefaultPostgresUser
	} else {
		// set env for pg_dump/pg_dumpall
		session.sh.SetEnv(EnvPgPassword, must(meta_util.GetBytesForKeys(appBindingSecret.Data, core.BasicAuthPasswordKey, envPostgresPassword)))
		userName = must(meta_util.GetBytesForKeys(appBindingSecret.Data, core.BasicAuthUsernameKey, envPostgresUser))
	}

	pgSSlmode, err := getSSLMODE(appBinding)
	if err != nil {
		return err
	}
	// Only set "PGSSLMODE" mode env variable, if it has been provided in the AppBinding.
	if pgSSlmode != "" {
		session.sh.SetEnv(EnvPGSSLMODE, pgSSlmode)
	}

	session.cmd.Args = append(session.cmd.Args, fmt.Sprintf("--username=%s", userName))
	return nil
}

func (session *sessionWrapper) setDatabaseConnectionParameters(appBinding *appcatalog.AppBinding) error {
	hostname, err := appBinding.Hostname()
	if err != nil {
		return err
	}
	session.cmd.Args = append(session.cmd.Args, fmt.Sprintf("--host=%s", hostname))

	if appBinding.Spec.ClientConfig.Service.Port == 0 {
		appBinding.Spec.ClientConfig.Service.Port = 5432
	}

	port, err := appBinding.Port()
	if err != nil {
		return err
	}
	session.cmd.Args = append(session.cmd.Args, fmt.Sprintf("--port=%d", port))

	return nil
}

func (session *sessionWrapper) setUserArgs(args string) {
	for _, arg := range strings.Fields(args) {
		session.cmd.Args = append(session.cmd.Args, arg)
	}
}

func (session *sessionWrapper) setTLSParameters(appBinding *appcatalog.AppBinding, scratchDir string) error {
	if appBinding.Spec.ClientConfig.CABundle != nil {
		if err := ioutil.WriteFile(filepath.Join(scratchDir, core.ServiceAccountRootCAKey), appBinding.Spec.ClientConfig.CABundle, 0600); err != nil {
			return err
		}
		session.sh.SetEnv(EnvPGSSLROOTCERT, filepath.Join(scratchDir, core.ServiceAccountRootCAKey))

	}
	return nil
}

func (session *sessionWrapper) waitForDBReady(waitTimeout int32) error {
	klog.Infoln("Waiting for the database to be ready.....")

	args := append(session.cmd.Args, fmt.Sprintf("--timeout=%d", waitTimeout))

	return session.sh.Command("pg_isready", args...).Run()
}

func getSSLMODE(appBinding *v1alpha1.AppBinding) (string, error) {
	sslmodeString := appBinding.Spec.ClientConfig.Service.Query
	if sslmodeString == "" {
		return "", nil
	}
	temps := strings.Split(sslmodeString, "=")
	if len(temps) != 2 {
		return "", fmt.Errorf("the sslmode is not valid. please provide the valid template. the temlpate should be like this: sslmode=<your_desire_sslmode>")
	}

	return strings.TrimSpace(temps[1]), nil
}
