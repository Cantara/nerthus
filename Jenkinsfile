def vers
def outFile
def release = false
pipeline {
    agent any
    tools {
        go 'Go 1.17'
        maven 'Mvn'
    }
    environment {
        NEXUS_CREDS = credentials('Cantara-NEXUS')
    }
    stages {
        stage("pre") {
            steps {
                script {
                    if (env.TAG_NAME) {
                        vers = "${env.TAG_NAME}"
                        release = true
                    } else {
                        vers = "${env.GIT_COMMIT}"
                    }
                    outFile = "nerthus-${vers}"
                    echo "New file: ${outFile}"
                }
            }
        }
        stage("test") {
            steps {
                script {
                    testApp()
                }
            }
        }
        stage("build") {
            steps {
                script {
                    echo "V: ${vers}"
                    echo "File: ${outFile}"
                    buildApp(outFile, vers)
                }
            }
        }
        stage("deploy") {
            steps {
                script {
                    echo 'deplying the application...'
                    echo "deploying version ${vers}"
                    if (release) {
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file ${outFile} https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/${outFile}"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/index.html https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/index.html"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/global.css https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/global.css"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/favicon.png https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/favicon.png"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.js https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.js"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.js.map https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.js.map"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.css https://mvnrepo.cantara.no/content/repositories/releases/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.css"
                    } else {
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file ${outFile} https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/${outFile}"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/index.html https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/index.html"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/global.css https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/global.css"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/favicon.png https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/favicon.png"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.js https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.js"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.js.map https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.js.map"
                        sh 'curl -v -u $NEXUS_CREDS '+"--upload-file frontend/public/build/bundle.css https://mvnrepo.cantara.no/content/repositories/snapshots/no/cantara/gotools/nerthus/${vers}/frontend/build/bundle.css"
                    }
                    sh "rm ${outFile}"
                    sh "rm -r frontend/npm"
                }
            }
        }
    }
}

def testApp() {
    echo 'testing the application...'
    sh './testRecursive.sh'
}

def buildApp(outFile, vers) {
    echo 'building the application...'
    sh 'ls'
    sh "CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -ldflags \"-X 'main.Version=${vers}' -X 'main.BuildTime=\$(date)'\" -o ${outFile}"
    sh 'cd frontend && mvn compile'
    sh 'ls'
}
