/**------------------------------------------------------------**
 * @filename project/xxx.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/5 17:11
 * @desc     go.jd100.com - main - summary
 **------------------------------------------------------------**/
package project

const (
	_tplDODockerfile = `FROM 172.16.0.185:5000/ubuntu:18.04.1
RUN mkdir -p /usr/src/app && cd /usr/src/app
COPY bin/{{.Name}}-{{.Module}} /usr/src/app/
EXPOSE 80
ENV DEPLOY_ENV pro
WORKDIR /usr/src/app/
ENTRYPOINT ./{{.Name}}-{{.Module}}

`
	_tplDODeploySh = `#!/bin/bash
cd {{.Name}}-{{.Module}}
PROJECT_NAME="{{.Name}}-{{.Module}}"
VERSION="1.0"
DEV_DOCKER_REGISTRY="172.16.0.185:5000"
ALI_DOCKER_REGISTRY="registry.cn-beijing.aliyuncs.com/jd100"
DEV_SERVER="172.16.0.252"   #开发部署服务器
TEST_SERVER="172.16.0.252"  #测试部署服务器

echo "版本："$VERSION

` + "commitid=`git rev-parse HEAD`" + `

echo "获取git 最新提交的sha作为tag一部分："$commitid
GITTAG=${commitid:0:8}

filename="../../../../"$PROJECT_NAME.$VERSION".count"
if [ ! -f $filename ];then
	cnt=1
else
	` + "cnt=`cat $filename`" + `
	` + "`expr $cnt + 1`" + `
fi
echo $cnt > $filename
echo "构建次数："$cnt

IMAGENAME=$VERSION"."$cnt"."$GITTAG
IMAGETAGE=$DEV_DOCKER_REGISTRY"/"$PROJECT_NAME":"$VERSION"."$cnt"."$GITTAG
echo "docker_image_name："$IMAGENAME

echo "start build..."
make all

echo "生成Docker 镜像...."
docker build -t $IMAGENAME .

echo "清除临时文件..."
make clean

echo "dev镜像打tag并push镜像"
docker image tag $IMAGENAME $IMAGETAGE
docker push $IMAGETAGE

if [ "$1" = 'ali' ];then
  echo "ali镜像打tag并push镜像"
  IMAGETAGE=$ALI_DOCKER_REGISTRY"/"$PROJECT_NAME":"$VERSION"."$cnt"."$GITTAG
  docker image tag $IMAGENAME $IMAGETAGE
  docker push $IMAGETAGE
fi

modConfig() {
  sed -i "s/\($2 *: *\).*/\1$3/" $1
}
echo "修改配置文件values.yaml和Chart.yaml"
modConfig "../../helm/"$PROJECT_NAME"/values.yaml" tag $IMAGENAME
modConfig "../../helm/"$PROJECT_NAME"/Chart.yaml" version $VERSION"."$cnt

server=$DEV_SERVER
if [ "$2" = 'test' ];then
  server=$TEST_SERVER
fi
echo "上传配置文件values.yaml和Chart.yaml到"$server
sshpass -p ubuntu scp -r "../../helm/"$PROJECT_NAME"/" "ubuntu@"$server":/home/ubuntu/helm-app/"$PROJECT_NAME"/"

echo "执行升级命令启动docker"
sshpass -p ubuntu ssh ubuntu@"$server" "helm upgrade "$PROJECT_NAME" /home/ubuntu/helm-app/"$PROJECT_NAME

`
	_tplDOJenkinsfile = `pipeline {
  agent any
  options {
    //不允许并行执行pipeline
    disableConcurrentBuilds()
    timeout(time: 15, unit: "MINUTES")
  }
  stages {
    stage('sonarQube代码检查') {
        steps {
            echo "starting codeAnalyze with SonarQube......"
              echo "任务名：$JOB_NAME"
              echo "分支名：${env.BRANCH_NAME}"
              script {
                scannerHome = tool name: 'sonar', type: 'hudson.plugins.sonar.SonarRunnerInstallation'
                projectName = JOB_NAME.replaceAll("/", "_")
              }
              echo "任务名：$projectName"
        }
    }

    stage('构建部署到内网开发服务器') {
      when {
        expression {
          script {
            lastLog = sh returnStdout: true, script: 'git log --pretty=format:"%s" -1'
            println env.BRANCH_NAME
          }
          return lastLog ==~ /.*build:dev.*/ || env.BRANCH_NAME ==~ /dev.*/
        }
      }
      agent any
      steps {
        echo "Deploying: $lastLog"
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh'
      }
    }

    stage('构建部署到内网测试服务器') {
      when {
        expression {
          script {
            lastLog = sh returnStdout: true, script: 'git log --pretty=format:"%s" -1'
          }
          return lastLog ==~ /.*build:test.*/ || env.BRANCH_NAME ==~ /test.*/
        }
      }
      agent any
      steps {
        echo "Deploying: $lastLog"
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh noAli test'
      }
    }

    stage('构建部署到内网测试服务器和阿里云，准备发布') {
      when {
        expression {
          script {
            lastLog = sh returnStdout: true, script: 'git log --pretty=format:"%s" -1'
          }
          return lastLog ==~ /.*build:ali.*/ || env.BRANCH_NAME ==~ /tag.*/
        }
      }
      agent any
      steps {
        echo "Deploying: $lastLog"
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh ali test'
      }
    }
  }
  post {
    failure {
      emailext(
        subject: "Jenkins build is ${currentBuild.result}: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
        mimeType: "text/html",
        body: """<p>Jenkins build is ${currentBuild.result}: ${env.JOB_NAME} #${env.BUILD_NUMBER}:</p>
            <p>Check jenkins console output at <a href="${env.BUILD_URL}console">${env.JOB_NAME} #${env.BUILD_NUMBER}</a></p>
            <p>Check sonarQube code analysis result at <a href="http://192.168.10.191:9000/dashboard?id=${projectName}">project ${projectName}</a></p>""",
        recipientProviders: [[$class: 'CulpritsRecipientProvider'],
                             [$class: 'DevelopersRecipientProvider'],
                             [$class: 'RequesterRecipientProvider']]
      )
    }
    success {
      emailext(
        subject: "Jenkins build is success: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
        mimeType: "text/html",
        body: """<p>Jenkins build is success: ${env.JOB_NAME} #${env.BUILD_NUMBER}:</p>
            <p>Check jenkins console output at <a href="${env.BUILD_URL}console">${env.JOB_NAME} #${env.BUILD_NUMBER}</a></p>
            <p>Check sonarQube code analysis result at <a href="http://192.168.10.191:9000/dashboard?id=${projectName}">project ${projectName}</a></p>""",
        recipientProviders: [[$class: 'CulpritsRecipientProvider'],
                             [$class: 'DevelopersRecipientProvider'],
                             [$class: 'RequesterRecipientProvider']]
      )
    }
  }
}

`
	_tplDOJenkinsfileA = `pipeline {
  agent none
  stages {
    stage('构建部署到内网测试服务器和阿里云，准备发布') {
      agent any
      steps {
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh ali test'
      }
    }
  }
}

`
	_tplDOJenkinsfileD = `pipeline {
  agent none
  stages {
    stage('构建部署到内网开发服务器') {
      agent any
      steps {
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh'
      }
    }
  }
}

`
	_tplDOJenkinsfileS = `pipeline {
  agent none
  stages {
    stage('sonarQube代码检查') {
        steps {
            echo "starting codeAnalyze with SonarQube......"
              echo "任务名：$JOB_NAME"
              echo "分支名：${env.BRANCH_NAME}"
              script {
                scannerHome = tool name: 'sonar', type: 'hudson.plugins.sonar.SonarRunnerInstallation'
                projectName = JOB_NAME.replaceAll("/", "_")
              }
              echo "任务名：$projectName"

            echo "scannerHome：${scannerHome}"
            withSonarQubeEnv('mySonarServer') {
              //注意这里withSonarQubeEnv()中的参数要与之前SonarQube servers中Name的配置相同
              sh "${scannerHome}/bin/sonar-scanner -X -Dsonar.language=php \
                                                    -Dsonar.projectKey=$projectName \
                                                    -Dsonar.projectName=$projectName \
                                                    -Dsonar.sources=./ \
                                                    -Dsonar.exclusions=vendor/**,install/**,src/backend/vendor/**,src/front/** \
                                                    -Dsonar.sourceEncoding=UTF-8"
            }
            script {
              timeout(2) {
                  //这里设置超时时间2分钟，不会出现一直卡在检查状态
                  //利用sonar webhook功能通知pipeline代码检测结果，未通过质量阈，pipeline将会fail
                  def qg = waitForQualityGate('mySonarServer')
                  //注意：这里waitForQualityGate()中的参数也要与之前SonarQube servers中Name的配置相同
                  if (qg.status != 'OK') {
                      error "未通过Sonarqube的代码质量阈检查，请及时修改！failure: ${qg.status}"
                  }
              }
            }
        }
    }
  }
}

`
	_tplDOJenkinsfileT = `pipeline {
  agent none
  stages {
    stage('构建部署到内网测试服务器') {
      agent any
      steps {
        sh 'chmod 777 install/build/deploy_docker.sh'
        sh 'install/build/deploy_docker.sh noAli test'
      }
    }
  }
}

`
	_tplDOMakefile = `all: build

build:
	cp -rf ../../../bin conf
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/{{.Name}}-{{.Module}} -tags static -ldflags -v ../../../{{.Name}}/cmd/main.go
clean:
	rm  -rf ./bin/
	rm -rf ./conf/

gotool:
	gofmt -w .

help:
	@echo "make - compile the source code"
	@echo "make clean - remove binary file and vim swp files"
	@echo "make gotool - run go tool 'fmt' and 'vet'"

.PHONY: clean gotool help...

`

	_tplDOHelmChart = `apiVersion: v1
appVersion: "1.0"
description: {{.Name}}-{{.Module}}服务
name: {{.Name}}-{{.Module}}
version: 1.0.0
`
	_tplDOHelmConfigMap = `apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap-{{.Name}}-{{.Module}}
  namespace: default
data:
  app.toml: |-
    name    = "{{.Name}}-{{.Module}}"
	version = "1.0.0"
	port    = ":80"
	appID   = 1
	# log setting default output stderr with json format.
	[log]
		level = "info"
		filters = ["instance_id", "zone"]
	# mysql database setting.
	[mysql]
		addr = "127.0.0.1:3306"
		dsn = "{user}:{password}@tcp(127.0.0.1:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8"
		readDSN = ["{user}:{password}@tcp(127.0.0.2:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8","{user}:{password}@tcp(127.0.0.3:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8,utf8mb4"]
		active = 20
		idle = 10
		idleTimeout ="4h"
		queryTimeout = "200ms"
		execTimeout = "300ms"
		tranTimeout = "400ms"
	# cache - redis setting.
	redisExpire = "24h"
	[redis]
		name = "{{.Name}}-{{.Module}}"
		proto = "tcp"
		addr = "127.0.0.1:6379"
		password = ""
		db = 8
		idle = 100
		active = 100
		dialTimeout = "1s"
		readTimeout = "1s"
		writeTimeout = "1s"
		idleTimeout = "10s"
	# mq - rabbit mq setting.
	[mq]
		dsn = "amqp://{user}:{password}@{host}:5672/{vhost}"
		[mq.exchange]
			name = "{exchange_name}"
			type = "{type}"
			routingKey = "{routing_key}"
			declare = true
			durable = true
			autoDelete = false
			internal = false
			noWait = false
			[mq.exchange.queue]
				 name = "{queue_name}"
	# rpc - grpc setting.
	[rpc.g]
		addr = "0.0.0.0:9000"
		timeout = "1s"

`
	_tplDOHelmValues = `# Default values for.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 1

name: {{.Name}}-{{.Module}}
image:
  repository: 172.16.0.185:5000/{{.Name}}-{{.Module}}
  tag: 1.0.0.d8575b19
  pullPolicy: IfNotPresent

nameOverride: ""
fullnameOverride: ""

service:
  namespace: default
  type: ClusterIP
  port: 80

ingress:
  enabled: true
  annotations: {}
    # kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  paths: []
  hosts:
    - {{.Name}}-{{.Module}}.jd
  tls: []
  #  - secretName: chart-example-tls
  #    hosts:
  #      - chart-example.local

resources: {}
  # We usually recommend not to specify default resources and to leave this as a conscious
  # choice for the user. This also increases chances charts run on environments with little
  # resources, such as Minikube. If you do want to specify resources, uncomment the following
  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
  # limits:
  #  cpu: 100m
  #  memory: 128Mi
  # requests:
  #  cpu: 100m
  #  memory: 128Mi
  
nodeSelector: {}
tolerations: []

affinity: {}

`
	_tplDOTemplateD = `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
  namespace: {{ .Values.service.namespace }}
  labels:
    app: {{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    metadata:
      labels:
        app: {{ .Chart.Name }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: {{ .Values.service.port }}
        resources:
          limits:
            cpu: "0.5"
            memory: 256Mi
        volumeMounts:
        - name: {{ .Chart.Name }}-app-config
          mountPath: /usr/src/app/conf
      volumes:
      - name: {{ .Chart.Name }}-app-config
        configMap:
           name: configmap-{{ .Chart.Name }}

`
	_tplDOTemplateI = `{{- if .Values.ingress.enabled -}}
---
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: gateway-{{ .Chart.Name }}
  namespace: {{ .Values.service.namespace }}
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    {{- range .Values.ingress.hosts }}
      - {{ . | quote }}
    {{- end }}
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}
  namespace: {{ .Values.service.namespace }}
spec:
  hosts:
  {{- range .Values.ingress.hosts }}
    - {{ . | quote }}
  {{- end }}
  gateways:
  - gateway-{{ .Chart.Name }}
  http:
  - route:
    - destination:
        host: {{ .Chart.Name }}
        port:
          number: {{ .Values.service.port }}
{{- end }}

`
	_tplDOTemplateS = `apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}
  namespace: {{ .Values.service.namespace }}
  labels:
    app: {{ .Chart.Name }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      name: http
  selector:
    app: {{.Chart.Name }}

`
)
