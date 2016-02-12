package henchman

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func moduleTestSetup(modName string) (module Module) {
	moduleContent := `
#!/usr/bin/env sh
ls -al $1
`
	writeTempFile([]byte(moduleContent), modName)

	mod, _ := NewModule(modName, "")
	return mod
}

func moduleTestTeardown(mod Module) {
	os.Remove(path.Join("/tmp", mod.Name))
}

func TestValidModule(t *testing.T) {
	name := "shell"
	args := "cmd=\"ls -al\" foo=bar baz=☃"
	mod, err := NewModule(name, args)

	require.NoError(t, err)

	assert.Equal(t, name, mod.Name, "Module names should match")
	assert.Equal(t, "ls -al", mod.Params["cmd"], "Mod params wasn't initialized properly")
	assert.Equal(t, "bar", mod.Params["foo"], "Expected value for foo to be bar")
	assert.Equal(t, "☃", mod.Params["baz"], "Expected a snowman")
}

func TestValidModuleWithVariousChars(t *testing.T) {
	name := "shell"
	args := "cmd='ls -al' url=\"http://foo.com/abc\" baz='abc-def' test=\"sed -i 's,store.enable=true,store.enable=false,g'\""
	mod, err := NewModule(name, args)

	require.NoError(t, err)

	assert.Equal(t, name, mod.Name, "Module names should match")
	assert.Equal(t, "ls -al", mod.Params["cmd"], "Mod params wasn't initialized properly")
	assert.Equal(t, "http://foo.com/abc", mod.Params["url"], "Expected value for foo to be bar")
	assert.Equal(t, "abc-def", mod.Params["baz"], "Expected a snowman")
	assert.Equal(t, "sed -i 's,store.enable=true,store.enable=false,g'", mod.Params["test"], "Expected 's,store.enable=true,store.enable=false,g'")

}


func TestValidModuleWithQuotes(t *testing.T) {
	name := "shell"
	args := "cmd=\"echo 'foo bar'\""
	mod, err := NewModule(name, args)
	require.NoError(t, err)

	assert.Equal(t, name, mod.Name, "Module names should match")
	assert.Equal(t, "echo 'foo bar'", mod.Params["cmd"], "Expected \"echo 'foo bar\"")
}


func TestInvalidArgsModule(t *testing.T) {
	name := "invalid"
	args := "foo"
	_, err := NewModule(name, args)

	require.Error(t, err, "Module arg parsing should have failed")
}

func TestInvalidArgsModule2(t *testing.T) {
	name := "invalid"
	args := "foo bar=baz"
	_, err := NewModule(name, args)

	require.Error(t, err, "Module arg parsing should have failed")
}

func TestModuleResolve(t *testing.T) {
	origSearchPath := ModuleSearchPath
	modDir := createTempDir("henchman")
	ModuleSearchPath = append(ModuleSearchPath, modDir)
	defer func() {
		ModuleSearchPath = origSearchPath
	}()
	defer os.RemoveAll(modDir)

	shellPath := path.Join(modDir, "shell")
	err := os.Mkdir(shellPath, 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(shellPath, "shell.linux"), []byte("ls -al"), 0644)
	mod, err := NewModule("shell", "foo=bar")

	require.NoError(t, err)
	require.NotNil(t, mod)

	fullPath, standalone, err := mod.Resolve("linux")

	require.NoError(t, err)
	assert.Equal(t, true, standalone)
	assert.Equal(t, path.Join(shellPath, "shell.linux"), fullPath, "Got incorrect fullPath")

	curlPath := path.Join(modDir, "curl", "curl")
	err = os.MkdirAll(curlPath, 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(curlPath, "exec"), []byte("ls -al"), 0644)
	mod, err = NewModule("curl", "foo=bar")

	require.NoError(t, err)
	require.NotNil(t, mod)

	fullPath, standalone, err = mod.Resolve("linux")

	require.NoError(t, err)
	assert.Equal(t, false, standalone)
	assert.Equal(t, curlPath, fullPath, "Got incorrect fullPath")
}

func setupTestShellModule() (Module, error) {
	writeTempFile([]byte("ls -al"), "shell")
	defer rmTempFile("/tmp/shell")
	return NewModule("shell", "foo=bar")
}

func TestNonexistentModuleResolve(t *testing.T) {
	//ModuleSearchPath = append(ModuleSearchPath, "/tmp")
	mod, err := setupTestShellModule()

	require.NoError(t, err)
	require.NotNil(t, mod)

	fullPath, _, err := mod.Resolve("linux")

	require.Error(t, err)
	require.Equal(t, "", fullPath, "Fullpath should have been empty")
}

func TestModuleDefaultExecOrder(t *testing.T) {
	mod, err := setupTestShellModule()

	require.NoError(t, err)
	require.NotNil(t, mod)

	require.NoError(t, InitConfiguration("../conf.json"))

	execOrder, err := mod.ExecOrder()
	require.NoError(t, err)

	assert.Equal(t, "exec_module", execOrder[0], "Exec Order sequence is wrong for a default module")
}

func TestModuleCopyExecOrder(t *testing.T) {
	writeTempFile([]byte("ls -al"), "copy")
	defer rmTempFile("/tmp/copy")
	mod, err := NewModule("copy", "src=foo dest=bar")

	require.NoError(t, err)
	require.NotNil(t, mod)

	require.NoError(t, InitConfiguration("../conf.json"))

	execOrder, err := mod.ExecOrder()
	require.NoError(t, err)

	assert.Equal(t, "stage", execOrder[0], "Exec Order sequence is wrong for copy module")
	assert.Equal(t, "exec_module", execOrder[1], "Exec Order sequence is wrong for copy module")
}

func TestModuleTemplateExecOrder(t *testing.T) {
	writeTempFile([]byte("ls -al"), "template")
	defer rmTempFile("/tmp/template")
	mod, err := NewModule("template", "src=foo dest=bar")

	require.NoError(t, err)
	require.NotNil(t, mod)

	require.NoError(t, InitConfiguration("../conf.json"))

	execOrder, err := mod.ExecOrder()
	require.NoError(t, err)

	assert.Equal(t, "process_template", execOrder[0], "Exec Order sequence is wrong for template module")
	assert.Equal(t, "stage", execOrder[1], "Exec Order sequence is wrong for template module")
	assert.Equal(t, "reset_src", execOrder[2], "Exec Order sequence is wrong for template module")
	assert.Equal(t, "exec_module", execOrder[3], "Exec Order sequence is wrong for template module")
}
