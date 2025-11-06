pipeline {
    agent {
        docker {
            image 'tmaier/docker-compose:latest'
            args '-v /var/run/docker.sock:/var/run/docker.sock'
        }
    }
    environment{
            GITLAB_USER = credentials('gitlab-user')
            GITLAB_ACCESS_TOKEN = credentials('gitlab-token')
            TELE_CHAT_ID = "-1001421406352"
            TELE_GITLABCI_TOKEN = credentials('telegram-token')
    }
    stages {
        stage('Test') {
            steps {
                sh 'apk add curl make git'
                sh 'docker-compose down'
                sh 'docker-compose --version'
                sh 'docker-compose up --build --exit-code-from tests'
                sh 'docker login -u ${GITLAB_USER} -p ${GITLAB_ACCESS_TOKEN} registry.gitlab.com'
                sh 'build-dev.sh registry.gitlab.com/wallet-gpay/orders-system'
            }
        }
    }
    post {
        failure {
           sh 'make telegram-failure'
        }
        success {
           sh 'make telegram-success'
        }
    }
}