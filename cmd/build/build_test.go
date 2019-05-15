package main

import (
	"os"
	"testing"
)

func TestBuild_BadDockerHost(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKER_HOST", "abc-123")
	main()
}

func TestBuild(t *testing.T) {
	os.Clearenv()
	main()
}
