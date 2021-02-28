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
	EnvPgPassword = "PGPASSWORD"
	PgDumpFile    = "dumpfile.sql"
	PgDumpCMD     = "pg_dump"
	PgDumpallCMD  = "pg_dumpall"
	PgRestoreCMD  = "psql"

	// Deprecated
	envPostgresUser = "POSTGRES_USER"
	// Deprecated
	envPostgresPassword = "POSTGRES_PASSWORD"
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

func waitForDBReady(appBinding *v1alpha1.AppBinding, secret *core.Secret, waitTimeout int32) error {
	log.Infoln("Waiting for the database to be ready.....")
	shell := sh.NewSession()
	shell.SetEnv(EnvPgPassword, must(meta_util.GetBytesForKeys(secret.Data, core.BasicAuthPasswordKey, envPostgresPassword)))
	args := []interface{}{
		fmt.Sprintf("--host=%s", appBinding.Spec.ClientConfig.Service.Name),
		fmt.Sprintf("--port=%d", appBinding.Spec.ClientConfig.Service.Port),
		fmt.Sprintf("--username=%s", must(meta_util.GetBytesForKeys(secret.Data, core.BasicAuthUsernameKey, envPostgresUser))),
		fmt.Sprintf("--timeout=%d", waitTimeout),
	}
	return shell.Command("pg_isready", args...).Run()
}
