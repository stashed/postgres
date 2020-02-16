/*
Copyright The Stash Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"fmt"
	"os/exec"
	"time"

	"stash.appscode.dev/apimachinery/pkg/restic"

	"github.com/appscode/go/log"
	"k8s.io/client-go/kubernetes"
	appcatalog_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD"
	EnvPgPassword    = "PGPASSWORD"
	PgDumpFile       = "dumpfile.sql"
	PgDumpCMD        = "pg_dumpall"
	PgRestoreCMD     = "psql"
)

type postgresOptions struct {
	kubeClient    kubernetes.Interface
	catalogClient appcatalog_cs.Interface

	namespace      string
	appBindingName string
	pgArgs         string
	outputDir      string

	setupOptions  restic.SetupOptions
	backupOptions restic.BackupOptions
	dumpOptions   restic.DumpOptions
}

func waitForDBReady(host string, port int32) {
	log.Infoln("Checking database connection")
	cmd := fmt.Sprintf(`nc "%s" "%d" -w 30`, host, port)
	for {
		if err := exec.Command(cmd).Run(); err != nil {
			break
		}
		log.Infoln("Waiting... database is not ready yet")
		time.Sleep(5 * time.Second)
	}
}
