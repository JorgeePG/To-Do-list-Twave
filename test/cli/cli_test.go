package cli_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommand_TextOutput(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "list", "--output", "text")
	cmd.Dir = "../../cmd"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output: %s", string(output))
	}
	assert.NoError(t, err)
	assert.NotEmpty(t, string(output))
}

func TestListCommand_JSONOutput(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "list", "--output", "json")
	cmd.Dir = "../../cmd"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output: %s", string(output))
	}
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(output), "[") || strings.HasPrefix(string(output), "null"))
}

func TestAddCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "add", "--title", "Tarea de prueba", "--finish", "false")
	cmd.Dir = "../../cmd"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output: %s", string(output))
	}
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Título: Tarea de prueba Tarea acabada: false")
}

func TestUserAddCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "user", "add", "--nombre", "Juan", "--atareado", "true")
	cmd.Dir = "../../cmd"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output: %s", string(output))
	}
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Nombre: Juan Tareas pendientes: true")
}

func TestUserDeleteCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "user", "delete", "--nombre", "Juan")
	cmd.Dir = "../../cmd"
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Output: %s", string(output))
	}
	assert.NoError(t, err)
	assert.Contains(t, string(output), "El usuario \" Juan \" ha sido eliminado")
}
