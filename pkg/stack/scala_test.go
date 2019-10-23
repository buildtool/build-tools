package stack

import (
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestScala_Scaffold_Error_Creating_Base_Directories(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, "test")
	_ = ioutil.WriteFile(filename, []byte("abc"), 0666)

	stack := &Scala{}

	err := stack.Scaffold(filename, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.EqualError(t, err, fmt.Sprintf("mkdir %s: not a directory", filename))
}

func TestScala_Scaffold_Error_Creating_Organisation_Directories(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "src", "main", "scala"), 0777)
	filename := filepath.Join(name, "src", "main", "scala", "org")
	_ = ioutil.WriteFile(filename, []byte("abc"), 0666)

	stack := &Scala{}

	err := stack.Scaffold(name, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.EqualError(t, err, fmt.Sprintf("mkdir %s: not a directory", filename))
}

func TestScala_Scaffold_Error_Creating_Dockerfile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "src", "main", "scala"), 0777)
	filename := filepath.Join(name, "project", "plugins.sbt")
	_ = os.MkdirAll(filename, 0777)

	stack := &Scala{}

	err := stack.Scaffold(name, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.EqualError(t, err, fmt.Sprintf("open %s: is a directory", filename))
}

func TestScala_Scaffold_Error_Appending_Dockerignore(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "src", "main", "scala"), 0777)
	filename := filepath.Join(name, ".dockerignore")
	_ = os.MkdirAll(filename, 0777)

	stack := &Scala{}

	err := stack.Scaffold(name, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.EqualError(t, err, fmt.Sprintf("open %s: is a directory", filename))
}

func TestScala_Scaffold(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(""), 0666)
	_ = ioutil.WriteFile(filepath.Join(name, ".gitignore"), []byte(""), 0666)
	stack := &Scala{}

	err := stack.Scaffold(name, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.NoError(t, err)
	assertFileContent(t, filepath.Join(name, "Dockerfile"), "FROM hseeberger/scala-sbt:11.0.3_1.3.0_2.13.0 as builder\nWORKDIR /build\n\nCOPY project /build/project\nRUN sbt update\n\nCOPY build.sbt /build\nRUN sbt update\n\nCOPY . /build\nRUN sbt clean coverage test coverageReport coverageOff stage\n\nFROM pliljenberg/java:latest\n\nWORKDIR /opt\n\nCOPY --from=builder /build/target/universal/stage/ /opt\n\nENTRYPOINT bin/test -J-Xmx\\$JVM_MEM\n\nCMD []\n")
	assertFileContent(t, filepath.Join(name, "build.sbt"), "lazy val logbackVersion         = \"1.2.3\"\nlazy val logbackJsonVersion     = \"5.2\"\nlazy val scalaTestVersion       = \"3.0.5\"\n\nlazy val `test` = (project in file(\".\")).\n  settings(\n    inThisBuild(List(\n      organization    := \"org.example\",\n      scalaVersion    := \"2.12.8\"\n    )),\n    name := \"test\",\n    resolvers ++= Seq(\n      Resolver.jcenterRepo\n    ),\n    libraryDependencies ++= Seq(\n      \"ch.qos.logback\"               % \"logback-classic\"          % logbackVersion,\n      \"net.logstash.logback\"         % \"logstash-logback-encoder\" % logbackJsonVersion,\n\n      \"org.scalatest\"               %% \"scalatest\"                % scalaTestVersion % Test\n    )\n  )\n\nenablePlugins(JavaAppPackaging)\n")
	assertFileContent(t, filepath.Join(name, "project", "plugins.sbt"), "addSbtPlugin(\"org.scalariform\" % \"sbt-scalariform\" % \"1.8.2\")\naddSbtPlugin(\"io.spray\" % \"sbt-revolver\" % \"0.9.1\")\naddSbtPlugin(\"com.typesafe.sbt\" % \"sbt-native-packager\" % \"1.3.15\")\naddSbtPlugin(\"org.scoverage\" % \"sbt-scoverage\" % \"1.5.1\")\n")
	assertFileContent(t, filepath.Join(name, "project", "build.properties"), "sbt.version=1.3.0\n")
	assertFileContent(t, filepath.Join(name, "src", "main", "resources", "logback.xml"), "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<configuration>\n  <appender name=\"TEXT\" class=\"ch.qos.logback.core.ConsoleAppender\">\n    <encoder>\n      <pattern>[%-5p] [%d{yyyy-MM-dd HH:mm:ss}] [%30.30logger{30}] %msg%n</pattern>\n    </encoder>\n  </appender>\n\n  <appender name=\"JSON\" class=\"ch.qos.logback.core.ConsoleAppender\">\n    <encoder class=\"net.logstash.logback.encoder.LogstashEncoder\" />\n  </appender>\n\n  <logger name=\"org.example\" level=\"DEBUG\" />\n\n  <root level=\"INFO\">\n    <appender-ref ref=\"${LOGFORMAT:-JSON}\" />\n  </root>\n</configuration>\n")
	assertFileContent(t, filepath.Join(name, ".dockerignore"), "\n\n.idea\ntarget\n*.iml\n")
	assertFileContent(t, filepath.Join(name, ".gitignore"), "\ntarget\n")
}

func TestScala_Name(t *testing.T) {
	stack := &Scala{}

	assert.Equal(t, "scala", stack.Name())
}
