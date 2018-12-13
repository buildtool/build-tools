#!/usr/bin/env bash

stack:scaffold:dockerfile() {
  local projectname="$1"
  cat <<EOF > Dockerfile
FROM nginx:1.13
COPY /files/* /usr/share/nginx/html/
COPY /default.conf /etc/nginx/conf.d/EOF
EOF
}

stack:scaffold:nginx:conf() {
  cat <<EOF > default.conf
server {
  listen       80;
  server_name  localhost;

  access_log /dev/stdout combined;
  error_log /dev/stdout info;

  location / {
    root   /usr/share/nginx/html;
    index  index.html index.htm;
    expires 0;
    try_files \$uri /index.html;
  }

  error_page   500 502 503 504  /50x.html;
  location = /50x.html {
    root   /usr/share/nginx/html;
  }
}
EOF
}

stack:scaffold:ingress() {
  local projectname="$1"
  local environment="$2"
  local domain="$3"
  local force_ssl="$4"
  local prefix="$environment-"
  prefix="${prefix#prod-}"

  cat <<EOF > "deployment_files/ingress-$environment.yaml"
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: ${projectname}-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "${force_ssl}"
spec:
  rules:
  - host: "${prefix}${projectname}.${domain}"
    http:
      paths:
      - path: /
        backend:
          serviceName: ${projectname}
          servicePort: 80
EOF
}

stack:scaffold:ingress:local() {
  local projectname="$1"
  local domain="$2"
  (stack:scaffold:ingress "$projectname" "local" "$domain" "false") >/dev/null
}

stack:scaffold:ingress:staging() {
  local projectname="$1"
  local domain="$2"
  (stack:scaffold:ingress "$projectname" "staging" "$domain" "true") >/dev/null
}

stack:scaffold:ingress:prod() {
  local projectname="$1"
  local domain="$2"
  (stack:scaffold:ingress "$projectname" "prod" "$domain" "true") >/dev/null
}

stack:validate() {
  local projectname="$1"
  : ${DOMAIN:?"DOMAIN must be set"}
}

stack:scaffold() {
  local projectname="$1"
  (stack:scaffold:dockerfile "$projectname") >/dev/null
  (stack:scaffold:nginx:conf) >/dev/null
  (stack:scaffold:ingress:local "$projectname" "$DOMAIN") >/dev/null
  (stack:scaffold:ingress:staging "$projectname" "$DOMAIN") >/dev/null
  (stack:scaffold:ingress:prod "$projectname" "$DOMAIN") >/dev/null
  mkdir -p files
  echo "$projectname" > files/index.html
}

stack:scaffold:dotfiles() {
  true
}
