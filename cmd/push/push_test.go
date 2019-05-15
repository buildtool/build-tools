package main

import (
	"os"
	"testing"
)

func TestPush_BadDockerHost(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKER_HOST", "abc-123")
	main()
}

func TestPush(t *testing.T) {
	os.Clearenv()
	main()
}
