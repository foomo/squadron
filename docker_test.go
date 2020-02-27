package configurd

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadService(t *testing.T) {
	file, err := os.Open("example/configurd/services/hello-service.yml")
	assert.NoError(t, err)
	defer file.Close()

	expected := Service{
		Name: "hello-service",
		Docker: Docker{
			File:    "Dockerfile",
			Context: "../application",
			Options: "",
			Image:   "foomo/configurd-hello",
		},
	}

	actual, err := LoadService(file)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestService_Build(t *testing.T) {
	svc := Service{
		Name: "hello-service",
		Docker: Docker{
			File:    "example/application/Dockerfile",
			Context: "example/application/",
			Options: "",
			Image:   "configurd-hello",
		},
	}

	output, err := svc.Build(context.Background(), "latest")
	assert.NoError(t, err)
	t.Log(output)
}
