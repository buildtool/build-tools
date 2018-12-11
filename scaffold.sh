#!/usr/bin/env bash

: ${DOCKER_REGISTRY_URL:?"DOCKER_REGISTRY_URL must be set"}

deployment:scaffold:create_deploy_file() {
  local projectname="$1"
  if [[ ! -f ./deployment_files/docker-compose.yml ]]
  then

    cat <<EOF > ./deployment_files/deploy.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: ${projectname}
  name: ${projectname}
  annotations:
    kubernetes.io/change-cause: "\${TIMESTAMP} Deployed commit id: \${COMMIT}"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ${projectname}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: ${projectname}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: "app"
                  operator: In
                  values:
                  - ${projectname}
              topologyKey: kubernetes.io/hostname
      containers:
      - name: ${projectname}
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 5
        imagePullPolicy: Always
        image: ${DOCKER_REGISTRY_URL}/${projectname}:\${COMMIT}
        ports:
        - containerPort: 80
      restartPolicy: Always
---

apiVersion: v1
kind: Service
metadata:
  name: ${projectname}
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: ${projectname}
  type: ClusterIP


EOF
  fi
}

scaffold:validate() {
  local projectname="$1"
  ci:validate "$projectname"
}

scaffold:mkdirs() {
  mkdir -p deployment_files
  ci:scaffold:mkdirs
}

scaffold:dotfiles() {
  touch .gitignore

  touch .editorconfig

  touch .dockerignore
  echo ".git" >> .dockerignore
  echo "Dockerfile" >> .dockerignore
  echo "README.md" >> .dockerignore
  echo ".editorconfig" >> .dockerignore

  ci:scaffold:dotfiles
  stack:scaffold:dotfiles
}

scaffold:create_readme() {
  local projectname="$1"
  local badges=$(ci:badges "$projectname")
  cat <<EOF >| README.md
# ${projectname}
${badges}
EOF
}

deployment:scaffold() {
  local projectname="$1"
  deployment:scaffold:create_deploy_file "$projectname"
}
