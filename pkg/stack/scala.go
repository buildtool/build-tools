package stack

import (
	"github.com/sparetimecoders/build-tools/pkg/file"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"os"
	"path/filepath"
	"strings"
)

type Scala struct{}

func (s Scala) Scaffold(dir string, data templating.TemplateData) error {
	for _, s := range []string{"main", "test"} {
		for _, t := range []string{"scala", "resources"} {
			if err := os.MkdirAll(filepath.Join(dir, "src", s, t), 0777); err != nil {
				return err
			}
		}
	}
	orgPath := append([]string{dir, "src", "main", "scala"}, strings.Split(data.Organisation, ".")...)
	if err := os.MkdirAll(filepath.Join(orgPath...), 0777); err != nil {
		return err
	}
	files := []struct {
		name    string
		content string
	}{
		{"Dockerfile", dockerfile},
		{"build.sbt", buildSbt},
		{"project/plugins.sbt", pluginsSbt},
		{"project/build.properties", "sbt.version=1.3.0"},
		{"src/main/resources/logback.xml", logbackXml},
	}
	for _, x := range files {
		if err := file.WriteTemplated(dir, x.name, x.content, data); err != nil {
			return err
		}
	}
	if err := file.Append(filepath.Join(dir, ".dockerignore"), dockerignore); err != nil {
		return err
	}
	return file.Append(filepath.Join(dir, ".gitignore"), "target")
}

func (s Scala) Name() string {
	return "scala"
}

var _ Stack = &Scala{}

var dockerignore = `
.idea
target
*.iml`

var dockerfile = `
FROM hseeberger/scala-sbt:11.0.3_1.3.0_2.13.0 as builder
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

ENTRYPOINT bin/{{.ProjectName}} -J-Xmx\$JVM_MEM

CMD []
`

var pluginsSbt = `
addSbtPlugin("org.scalariform" % "sbt-scalariform" % "1.8.2")
addSbtPlugin("io.spray" % "sbt-revolver" % "0.9.1")
addSbtPlugin("com.typesafe.sbt" % "sbt-native-packager" % "1.3.15")
addSbtPlugin("org.scoverage" % "sbt-scoverage" % "1.5.1")
`

var buildSbt = `
lazy val logbackVersion         = "1.2.3"
lazy val logbackJsonVersion     = "5.2"
lazy val scalaTestVersion       = "3.0.5"

lazy val ` + "`{{.ProjectName}}`" + ` = (project in file(".")).
  settings(
    inThisBuild(List(
      organization    := "{{.Organisation}}",
      scalaVersion    := "2.12.8"
    )),
    name := "{{.ProjectName}}",
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
`

var logbackXml = `
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

  <logger name="{{.Organisation}}" level="DEBUG" />

  <root level="INFO">
    <appender-ref ref="${LOGFORMAT:-JSON}" />
  </root>
</configuration>
`
