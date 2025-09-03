pipeline {
  agent any
  options { skipDefaultCheckout(true) }

  environment {
    FAIL_ON_ISSUES     = 'false'
    SONAR_HOST_URL     = 'http://sonarqube:9000'
    SONAR_PROJECT_KEY  = 'backend-api-golang'
    SONAR_PROJECT_NAME = 'backend-api-golang'
  }

  stages {
    stage('Clean Workspace') { steps { cleanWs() } }

    stage('Checkout') {
      steps {
        checkout([$class: 'GitSCM',
          branches: [[name: '*/main']],
          extensions: [[$class: 'CloneOption', shallow: false, noTags: false]],
          userRemoteConfigs: [[url: 'https://github.com/yosua789/go-sast-test.git']]
        ])
        sh 'echo "WS: $WORKSPACE" && ls -la'
      }
    }

    stage('Build & Test (Go)') {
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          sh '''
            set -eux
            if [ -f go.mod ]; then
              cat > .ci-go.sh <<'EOS'
#!/usr/bin/env bash
set -eux
go version
go env -w GOMODCACHE=/tmp/go-mod-cache GOPATH=/tmp/go
mkdir -p /tmp/go-mod-cache /tmp/go
go mod download
go build ./... || true
go test ./... -count=1 -covermode=atomic -coverprofile=coverage-go.out || true
go test -json ./... > gotest-report.json || true
EOS
              chmod +x .ci-go.sh
              docker run --rm --network jenkins --volumes-from jenkins -w "$WORKSPACE" --platform linux/arm64 golang:1.22-bookworm bash ./.ci-go.sh
            else
              echo "No go.mod, skip Go build."
            fi
          '''
        }
      }
    }

    stage('SAST - SonarQube') {
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          withCredentials([string(credentialsId: 'sonarqube-token', variable: 'T')]) {
            sh '''
              set -eux
              docker pull sonarsource/sonar-scanner-cli
              EXTRA_GO_FLAGS=""
              [ -f coverage-go.out ] && EXTRA_GO_FLAGS="$EXTRA_GO_FLAGS -Dsonar.go.coverage.reportPaths=coverage-go.out"
              docker run --rm --network jenkins \
                --volumes-from jenkins -w "$WORKSPACE" \
                -e SONAR_HOST_URL="$SONAR_HOST_URL" \
                sonarsource/sonar-scanner-cli \
                  -Dsonar.host.url="$SONAR_HOST_URL" \
                  -Dsonar.token="$T" \
                  -Dsonar.scanner.socketTimeout=300 \
                  -Dsonar.projectKey="$SONAR_PROJECT_KEY" \
                  -Dsonar.projectName="$SONAR_PROJECT_NAME" \
                  -Dsonar.scm.provider=git \
                  -Dsonar.sources=. \
                  -Dsonar.inclusions="**/*" \
                  -Dsonar.exclusions="**/log/**,**/node_modules/**,**/dist/**,**/build/**,**/target/**,**/vendor/**,**/.venv/**,**/__pycache__/**,**/*.pyc,docker-compose.yaml" \
                  -Dsonar.coverage.exclusions="**/*.test.*,**/test/**,**/tests**" \
                  ${EXTRA_GO_FLAGS}
            '''
          }
        }
      }
    }

    stage('SCA - Dependency-Check (Go)') {
      agent {
        docker {
          image 'owasp/dependency-check:latest'
          reuseNode true
          args "--entrypoint='' -e JVM_OPTS=-Xmx1024m"
        }
      }
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          sh '''
            set -euxo pipefail
            rm -rf dependency-check-report || true
            mkdir -p dependency-check-report .dc-data

            DATA_DIR="$WORKSPACE/.dc-data"
            SCANS=""
            [ -f go.mod ] && SCANS="$SCANS --scan go.mod"
            [ -f go.sum ] && SCANS="$SCANS --scan go.sum"
            [ -d vendor ] && SCANS="$SCANS --scan vendor"

            if [ -z "$SCANS" ]; then
              echo "<html><body><h1>No dependencies scanned</h1></body></html>" > dependency-check-report/dependency-check-report.html
            else
              set +e
              /usr/share/dependency-check/bin/dependency-check.sh --updateonly --data "$DATA_DIR" || true
              set -e
              /usr/share/dependency-check/bin/dependency-check.sh \
                --project "backend-api-golang" \
                $SCANS \
                --enableExperimental \
                --format HTML,JSON,JUNIT,SARIF \
                --out dependency-check-report \
                --log dependency-check-report/dependency-check.log \
                --data "$DATA_DIR" \
                --failOnCVSS 11 || true
            fi

            echo "== DC OUTPUT =="
            ls -lah dependency-check-report || true
          '''
        }
      }
      post {
        always {
          script {
            if (fileExists('dependency-check-report')) {
              archiveArtifacts artifacts: 'dependency-check-report/**', fingerprint: false, onlyIfSuccessful: false, allowEmptyArchive: true
            }
          }
          publishHTML(target: [
            reportName: 'Dependency-Check',
            reportDir:  'dependency-check-report',
            reportFiles:'dependency-check-report.html',
            keepAll: true,
            alwaysLinkToLastBuild: true,
            allowMissing: true
          ])
        }
      }
    }

    stage('SCA - Trivy (filesystem)') {
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          sh '''
            set -euxo pipefail
            rm -rf reports/trivy || true
            mkdir -p reports/trivy .trivy-cache
            chmod -R 777 .trivy-cache || true

            cat > reports/trivy/html.tpl <<'TPL'
<!DOCTYPE html><html><head><meta charset="utf-8"><title>Trivy FS Report</title>
<style>table{border-collapse:collapse}td,th{border:1px solid #999;padding:4px 8px}th{background:#eee}</style>
</head><body>
<h1>Trivy Filesystem Report</h1>
<p>Generated: {{ .Report.Meta.CreatedAt }}</p>
<table>
<tr><th>Target</th><th>VulnerabilityID</th><th>PkgName</th><th>Installed</th><th>Severity</th><th>Title</th></tr>
{{- range .Results }}
  {{- $t := .Target }}
  {{- range .Vulnerabilities }}
<tr>
<td>{{ $t }}</td>
<td>{{ .VulnerabilityID }}</td>
<td>{{ .PkgName }}</td>
<td>{{ .InstalledVersion }}</td>
<td>{{ .Severity }}</td>
<td>{{ .Title }}</td>
</tr>
  {{- end }}
{{- end }}
</table></body></html>
TPL

            docker pull aquasec/trivy:latest

            docker run --rm --network jenkins \
              --volumes-from jenkins -w "$WORKSPACE" \
              -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
              -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
              aquasec/trivy:latest \
              fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
              --severity HIGH,CRITICAL . | tee reports/trivy/trivy-fs.txt

            docker run --rm --network jenkins \
              --volumes-from jenkins -w "$WORKSPACE" \
              -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
              -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
              aquasec/trivy:latest \
              fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
              --severity HIGH,CRITICAL --format sarif -o reports/trivy/trivy-fs.sarif .

            docker run --rm --network jenkins \
              --volumes-from jenkins -w "$WORKSPACE" \
              -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
              -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
              -v "$WORKSPACE/reports/trivy:/out" \
              -v "$WORKSPACE/reports/trivy/html.tpl:/html.tpl:ro" \
              aquasec/trivy:latest \
              fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
              --severity HIGH,CRITICAL --format template --template "@/html.tpl" -o /out/index.html .

            echo "== TRIVY OUTPUT =="
            ls -lah reports/trivy || true
          '''
        }
      }
      post {
        always {
          script {
            if (fileExists('reports/trivy/trivy-fs.txt'))   archiveArtifacts artifacts: 'reports/trivy/trivy-fs.txt', fingerprint: false
            if (fileExists('reports/trivy/trivy-fs.sarif')) archiveArtifacts artifacts: 'reports/trivy/trivy-fs.sarif', fingerprint: false
            if (fileExists('reports/trivy/index.html'))     archiveArtifacts artifacts: 'reports/trivy/index.html', fingerprint: false
          }
          publishHTML(target: [
            reportName: 'Trivy Report',
            reportDir:  'reports/trivy',
            reportFiles:'index.html',
            keepAll: true,
            alwaysLinkToLastBuild: true,
            allowMissing: false
          ])
          catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
            recordIssues enabledForFailure: true,
              tool: sarif(pattern: 'reports/trivy/trivy-fs.sarif'),
              trendChartType: 'TOOLS_ONLY'
          }
        }
      }
    }

    stage('SAST - Semgrep') {
      agent {
        docker {
          image 'semgrep/semgrep:latest'
          reuseNode true
          args "--entrypoint=''"
        }
      }
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          script {
            sh '''
              set +e
              semgrep --version || true
              semgrep scan \
                --config p/ci --config p/owasp-top-ten --config p/docker \
                --exclude 'log/**' --exclude '**/node_modules/**' --exclude '**/dist/**' --exclude '**/build/**' \
                --severity ERROR --error \
                --sarif --output semgrep.sarif .
              semgrep scan \
                --config p/ci --config p/owasp-top-ten --config p/docker \
                --exclude 'log/**' --exclude '**/node_modules/**' --exclude '**/dist/**' --exclude '**/build/**' \
                --severity ERROR --error \
                --junit-xml --output semgrep-junit.xml .
              echo $? > .semgrep_exit
            '''
            def ec = readFile('.semgrep_exit').trim()
            if (env.FAIL_ON_ISSUES == 'true' && ec != '0') { error "Fail build (policy) Semgrep exit ${ec}" }
            sh 'ls -lh semgrep.* || true'
          }
        }
      }
      post {
        always {
          script {
            if (fileExists('semgrep.sarif'))      archiveArtifacts artifacts: 'semgrep.sarif', fingerprint: false
            if (fileExists('semgrep-junit.xml')) { archiveArtifacts artifacts: 'semgrep-junit.xml', fingerprint: false
              junit allowEmptyResults: true, testResults: 'semgrep-junit.xml', skipPublishingChecks: true, skipMarkingBuildUnstable: true }
          }
        }
      }
    }
  }

  post { always { echo "Scanning All Done. Result: ${currentBuild.currentResult}" } }
}
