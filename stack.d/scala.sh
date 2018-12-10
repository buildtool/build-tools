#!/usr/bin/env bash

: ${ORGANIZATION:?"ORGANIZATION must be set"}

stack:scaffold:dockerfile() {
  local projectname="$1"
  cat <<EOF > Dockerfile
FROM pliljenberg/scala-sbt:2.12.8-1.2.7 as builder
WORKDIR /build

COPY project /build/project
RUN sbt update

COPY build.sbt /build
RUN sbt update

COPY . /build
RUN sbt clean coverage test coverageReport coverageOff stage

FROM pliljenberg/java:latest

WORKDIR /opt

COPY --from=builder /build/target/universal/stage/ /opt

ENTRYPOINT bin/${projectname} -J-Xmx\$JVM_MEM

CMD []
EOF
}

stack:scala:scaffold() {
  local projectname="$1"

  mkdir -p src/{main,test}/{scala,resources}
  mkdir -p "src/main/scala/$(echo $ORGANIZATION | sed 's/\./\//')"
  mkdir project

  cat <<EOF > project/plugins.sbt
addSbtPlugin("org.scalariform" % "sbt-scalariform" % "1.8.2")
addSbtPlugin("io.spray" % "sbt-revolver" % "0.9.1")
addSbtPlugin("com.typesafe.sbt" % "sbt-native-packager" % "1.3.15")
addSbtPlugin("org.scoverage" % "sbt-scoverage" % "1.5.1")
EOF

  echo "sbt.version=1.2.7" > project/build.properties

  cat <<EOF > build.sbt
lazy val logbackVersion         = "1.2.3"
lazy val logbackJsonVersion     = "5.2"
lazy val scalaTestVersion       = "3.0.5"

lazy val \`$projectname\` = (project in file(".")).
  settings(
    inThisBuild(List(
      organization    := "$ORGANIZATION",
      scalaVersion    := "2.12.8"
    )),
    name := "$projectname",
    resolvers ++= Seq(
      Resolver.jcenterRepo
    ),
    libraryDependencies ++= Seq(
      "ch.qos.logback"               % "logback-classic"          % logbackVersion,
      "net.logstash.logback"         % "logstash-logback-encoder" % logbackJsonVersion,

      "org.scalatest"               %% "scalatest"                % scalaTestVersion % Test
    )
  )

enablePlugins(JavaAppPackaging)
EOF

    cat <<EOF > src/main/resources/logback.xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
  <appender name="TEXT" class="ch.qos.logback.core.ConsoleAppender">
    <encoder>
      <pattern>[%-5p] [%d{yyyy-MM-dd HH:mm:ss}] [%30.30logger{30}] %msg%n</pattern>
    </encoder>
  </appender>

  <appender name="JSON" class="ch.qos.logback.core.ConsoleAppender">
    <encoder class="net.logstash.logback.encoder.LogstashEncoder" />
  </appender>

  <logger name="$ORGANIZATION" level="DEBUG" />

  <root level="INFO">
    <appender-ref ref="\${LOGFORMAT:-JSON}" />
  </root>
</configuration>

EOF
}

stack:scaffold() {
  local projectname="$1"

  stack:scaffold:dockerfile "$projectname"
  stack:scala:scaffold "$projectname"
}

stack:scaffold:dotfiles() {
  echo ".idea" >> .gitignore
  echo "target" >> .gitignore
  echo "*.iml" >> .gitignore

  echo "target" >> .dockerignore
}
