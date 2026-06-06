import jenkins.model.*
import org.jenkinsci.plugins.workflow.job.WorkflowJob

def jenkins = Jenkins.getInstanceOrNull()
if (jenkins == null) return

def jobName = 'common-ci-pipelines'
def job = jenkins.getItem(jobName)

if (job == null) {
    job = jenkins.createProject(WorkflowJob.class, jobName)
    println "Created pipeline job: ${jobName}"
}

// 真实的 CI 流水线：拉代码 → 构建镜像 → 推送到本地 Registry
def pipelineScript = '''
pipeline {
    agent any

    parameters {
        string(name: 'GIT_REPO', defaultValue: 'https://jihulab.com/stevehailong-group/stevehailong-project.git', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: '分支')
        string(name: 'IMAGE_NAME', defaultValue: 'mycloud/common-ci-pipelines', description: '镜像名')
        string(name: 'REGISTRY', defaultValue: '172.18.0.1:5001', description: '镜像仓库地址')
        string(name: 'VERSION', defaultValue: '', description: '版本号（留空自动生成）')
    }

    stages {
        stage('Checkout') {
            steps {
                script {
                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${params.GIT_BRANCH}"]],
                        userRemoteConfigs: [[url: params.GIT_REPO]]
                    ])
                }
            }
        }

        stage('Build') {
            steps {
                script {
                    def version = params.VERSION
                    if (!version) {
                        version = "1.0.${env.BUILD_ID}-a1b2c3d"
                    }
                    def imageTag = "${params.REGISTRY}/${params.IMAGE_NAME}:${version}"

                    sh "docker build -t ${imageTag} ."
                    sh "docker push ${imageTag}"

                    currentBuild.description = "镜像: ${imageTag}"
                    println "Built and pushed: ${imageTag}"
                }
            }
        }
    }

    post {
        success { println 'CI Build Success' }
        failure { println 'CI Build Failed' }
    }
}
'''

job.setDefinition(new org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition(pipelineScript, true))
job.save()
println "Updated pipeline job script: ${jobName}"
