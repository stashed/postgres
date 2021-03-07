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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	stash "stash.appscode.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/pkg/restic"

	"github.com/codeskyblue/go-sh"
	"gomodules.xyz/x/log"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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

func (opt *postgresOptions) waitForDBReady(appBinding *v1alpha1.AppBinding, secret *core.Secret, waitTimeout int32) error {
	log.Infoln("Waiting for the database to be ready.....")
	shell := sh.NewSession()

	if appBinding.Spec.ClientConfig.Service.Port == 0 {
		appBinding.Spec.ClientConfig.Service.Port = 5432
	}

	userName := ""
	if _, ok := secret.Data[core.TLSPrivateKeyKey]; ok {
		certByte, ok := secret.Data[core.TLSCertKey]
		if !ok {
			return fmt.Errorf("can't find client cert")
		}
		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey), certByte, 0600); err != nil {
			return err
		}

		shell.SetEnv(EnvPGSSLCERT, filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey))
		keyByte, ok := secret.Data[core.TLSPrivateKeyKey]
		if !ok {
			return fmt.Errorf("can't find client private key")
		}

		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey), keyByte, 0600); err != nil {
			return err
		}
		shell.SetEnv(EnvPGSSLKEY, filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey))

		//TODO: this one is hard coded here but need to change later
		userName = DefaultPostgresUser
	} else {
		// set env for pg_dump/pg_dumpall
		shell.SetEnv(EnvPgPassword, must(meta_util.GetBytesForKeys(secret.Data, core.BasicAuthPasswordKey, envPostgresPassword)))
		userName = must(meta_util.GetBytesForKeys(secret.Data, core.BasicAuthUsernameKey, envPostgresUser))

	}

	if appBinding.Spec.ClientConfig.CABundle != nil {
		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.ServiceAccountRootCAKey), appBinding.Spec.ClientConfig.CABundle, 0600); err != nil {
			return err
		}
		shell.SetEnv(EnvPGSSLROOTCERT, filepath.Join(opt.setupOptions.ScratchDir, core.ServiceAccountRootCAKey))

	}
	pgSSlmode, err := getSSLMODE(appBinding)
	if err != nil {
		return err
	}
	shell.SetEnv(EnvPGSSLMODE, pgSSlmode)

	//shell.SetEnv(EnvPgPassword, must(meta_util.GetBytesForKeys(secret.Data, core.BasicAuthPasswordKey, envPostgresPassword)))
	args := []interface{}{
		fmt.Sprintf("--host=%s", appBinding.Spec.ClientConfig.Service.Name),
		fmt.Sprintf("--port=%d", appBinding.Spec.ClientConfig.Service.Port),
		fmt.Sprintf("--username=%s", userName),
		fmt.Sprintf("--timeout=%d", waitTimeout),
	}

	return shell.Command("pg_isready", args...).Run()
}

func getSSLMODE(appBinding *v1alpha1.AppBinding) (string, error) {

	sslmodeString := appBinding.Spec.ClientConfig.Service.Query
	temps := strings.Split(sslmodeString, "=")
	if len(temps) != 2 {
		return "", fmt.Errorf("the sslmode is not valid. please provide the valid template. the temlpate should be like this: sslmode=<your_desire_sslmode>")
	}

	return strings.TrimSpace(temps[1]), nil
}

func (opt *postgresOptions) GetResticWrapperWithPGConnectorVariables(appBinding *v1alpha1.AppBinding, appBindingSecret *core.Secret) (resticWrapper *restic.ResticWrapper, userName string, err error) {
	userName = ""
	resticWrapper, err = restic.NewResticWrapper(opt.setupOptions)
	if err != nil {
		return nil, userName, err
	}
	if _, ok := appBindingSecret.Data[core.TLSPrivateKeyKey]; ok {
		certByte, ok := appBindingSecret.Data[core.TLSCertKey]
		if !ok {
			return nil, userName, fmt.Errorf("can't find client cert")
		}
		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey), certByte, 0600); err != nil {
			return nil, userName, err
		}

		resticWrapper.SetEnv(EnvPGSSLCERT, filepath.Join(opt.setupOptions.ScratchDir, core.TLSCertKey))
		keyByte, ok := appBindingSecret.Data[core.TLSPrivateKeyKey]
		if !ok {
			return nil, userName, fmt.Errorf("can't find client private key")
		}

		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey), keyByte, 0600); err != nil {
			return nil, userName, err
		}
		resticWrapper.SetEnv(EnvPGSSLKEY, filepath.Join(opt.setupOptions.ScratchDir, core.TLSPrivateKeyKey))

		//TODO: this one is hard coded here but need to change later
		userName = DefaultPostgresUser
	} else {
		// set env for pg_dump/pg_dumpall
		resticWrapper.SetEnv(EnvPgPassword, must(meta_util.GetBytesForKeys(appBindingSecret.Data, core.BasicAuthPasswordKey, envPostgresPassword)))
		userName = must(meta_util.GetBytesForKeys(appBindingSecret.Data, core.BasicAuthUsernameKey, envPostgresUser))

	}

	if appBinding.Spec.ClientConfig.CABundle != nil {
		if err := ioutil.WriteFile(filepath.Join(opt.setupOptions.ScratchDir, core.ServiceAccountRootCAKey), appBinding.Spec.ClientConfig.CABundle, 0600); err != nil {
			return nil, userName, err
		}
		resticWrapper.SetEnv(EnvPGSSLROOTCERT, filepath.Join(opt.setupOptions.ScratchDir, core.ServiceAccountRootCAKey))

	}
	pgSSlmode, err := getSSLMODE(appBinding)
	if err != nil {
		return nil, userName, err
	}
	resticWrapper.SetEnv(EnvPGSSLMODE, pgSSlmode)
	return resticWrapper, userName, nil
}
