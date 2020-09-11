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

	"github.com/appscode/go/log"
	"github.com/codeskyblue/go-sh"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD"
	EnvPgPassword    = "PGPASSWORD"
	PgDumpFile       = "dumpfile.sql"
	PgDumpCMD        = "pg_dump"
	PgDumpallCMD     = "pg_dumpall"
	PgRestoreCMD     = "psql"
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

func waitForDBReady(appBinding *v1alpha1.AppBinding, secret *core.Secret, waitTimeout int32) error {
	log.Infoln("Waiting for the database to be ready.....")
	shell := sh.NewSession()
	shell.SetEnv(EnvPgPassword, string(secret.Data[PostgresPassword]))
	args := []interface{}{
		fmt.Sprintf("--host=%s", appBinding.Spec.ClientConfig.Service.Name),
		fmt.Sprintf("--port=%d", appBinding.Spec.ClientConfig.Service.Port),
		fmt.Sprintf("--username=%s", secret.Data[PostgresUser]),
		fmt.Sprintf("--timeout=%d", waitTimeout),
	}
	return shell.Command("pg_isready", args...).Run()
}
