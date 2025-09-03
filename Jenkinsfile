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
              docker run --rm --network jenkins --volumes-from jenkins -w "$WORKSPACE" golang:1.22-bullseye bash -lc '
                go env -w GOMODCACHE=/tmp/go-mod-cache GOPATH=/tmp/go
                mkdir -p /tmp/go-mod-cache /tmp/go
                go mod download
                go build ./... || true
                go test ./... -count=1 -covermode=atomic -coverprofile=coverage-go.out || true
                go test -json ./... > gotest-report.json || true
              '
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
                  -Dsonar.login="$T" \
                  -Dsonar.ws.timeout=120 \
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
          args "--entrypoint=''"
        }
      }
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          script {
            sh '''
              set -eux
              mkdir -p dependency-check-report
              /usr/share/dependency-check/bin/dependency-check.sh --updateonly || true
              set +e
              /usr/share/dependency-check/bin/dependency-check.sh \
                --project "backend-api-golang" \
                --scan go.mod --scan go.sum --scan vendor \
                --enableExperimental \
                --format ALL \
                --out dependency-check-report \
                --log dependency-check-report/dependency-check.log \
                --failOnCVSS 11
              echo $? > .dc_exit
            '''
            def rc = readFile('.dc_exit').trim()
            echo "Dependency-Check exit code: ${rc}"
            if (env.FAIL_ON_ISSUES == 'true' && rc != '0') {
              error "Fail build (policy) Dependency-Check exit ${rc}"
            }
          }
        }
      }
      post {
        always {
          script {
            if (fileExists('dependency-check-report')) {
              archiveArtifacts artifacts: 'dependency-check-report/**', fingerprint: false, onlyIfSuccessful: false
            }
          }
        }
      }
    }

    stage('SCA - Trivy (filesystem)') {
      steps {
        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
          script {
            sh '''
              set -eux
              rm -f trivy-fs.txt trivy-fs.sarif trivy-report.html trivy-html.tpl || true
              mkdir -p .trivy-cache
              chmod -R 777 .trivy-cache || true
              docker pull aquasec/trivy:latest

              docker run --rm --network jenkins --volumes-from jenkins -w "$WORKSPACE" \
                curlimages/curl:8.8.0 -sSL -o trivy-html.tpl https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/html.tpl

              docker run --rm --network jenkins \
                --volumes-from jenkins -w "$WORKSPACE" \
                -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
                -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
                aquasec/trivy:latest \
                fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
                   --severity HIGH,CRITICAL . | tee trivy-fs.txt

              docker run --rm --network jenkins \
                --volumes-from jenkins -w "$WORKSPACE" \
                -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
                -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
                aquasec/trivy:latest \
                fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
                   --severity HIGH,CRITICAL --format sarif -o trivy-fs.sarif .

              docker run --rm --network jenkins \
                --volumes-from jenkins -w "$WORKSPACE" \
                -u 0:0 -e HOME=/tmp -e XDG_CACHE_HOME=/tmp/trivy-cache \
                -v "$WORKSPACE/.trivy-cache:/tmp/trivy-cache:rw" \
                -v "$WORKSPACE/trivy-html.tpl:/trivy-html.tpl:ro" \
                aquasec/trivy:latest \
                fs --cache-dir /tmp/trivy-cache --no-progress --exit-code 0 \
                   --severity HIGH,CRITICAL \
                   --format template --template "@/trivy-html.tpl" \
                   -o trivy-report.html .
              echo 0 > .trivy_exit
            '''
            def ec = readFile('.trivy_exit').trim()
            echo "Trivy FS scan exit code: ${ec}"
            if (env.FAIL_ON_ISSUES == 'true' && ec != '0') {
              error "Fail build (policy) Trivy FS exit ${ec}"
            }
            sh 'ls -lh trivy-fs.* trivy-report.html || true'
          }
        }
      }
      post {
        always {
          script {
            if (fileExists('trivy-fs.txt'))      archiveArtifacts artifacts: 'trivy-fs.txt', fingerprint: false
            if (fileExists('trivy-fs.sarif'))    archiveArtifacts artifacts: 'trivy-fs.sarif', fingerprint: false
            if (fileExists('trivy-report.html')) archiveArtifacts artifacts: 'trivy-report.html', fingerprint: false
          }
          publishHTML(target: [
            reportName: 'Trivy Report',
            reportDir:  '.',
            reportFiles:'trivy-report.html',
            keepAll: true,
            alwaysLinkToLastBuild: true,
            allowMissing: true
          ])
          catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
            recordIssues enabledForFailure: true,
              tool: sarif(pattern: 'trivy-fs.sarif'),
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
            if (env.FAIL_ON_ISSUES == 'true' && ec != '0') {
              error "Fail build (policy) Semgrep exit ${ec}"
            }
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
