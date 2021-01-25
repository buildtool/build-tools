# TeamCity
[TeamCity] can be configured with a `.teamcity/settings.kts` file in your project.

```kotlin
import jetbrains.buildServer.configs.kotlin.v2018_2.*
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.script
import jetbrains.buildServer.configs.kotlin.v2018_2.triggers.finishBuildTrigger
import jetbrains.buildServer.configs.kotlin.v2018_2.triggers.vcs

version = "2019.1"

project {
    buildType(BuildAndPush)
}

object BuildAndPush : BuildType({
    name = "BuildAndPush"

    steps {
        script {
            name = "build and push"
            scriptContent = """
                build && push
            """.trimIndent()
            dockerImage = "buildtool/buildtools"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
            dockerPull = true
            dockerRunParameters = """
                -v /var/run/docker.sock:/var/run/docker.sock
                --rm
            """.trimIndent()
        }
    }

    triggers {
        vcs {}
    }
})

```

[teamcity]: https://www.jetbrains.com/teamcity
