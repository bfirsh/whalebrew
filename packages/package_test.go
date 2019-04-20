package packages

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func newTestPkg(label, value string) (*Package, error) {
	return NewPackageFromImage("whalebrew/whalesay", types.ImageInspect{
		ContainerConfig: &container.Config{
			Labels: map[string]string{label: value},
		},
	})
}

func mustNewTestPkg(t *testing.T, label, value string) *Package {
	pkg, err := newTestPkg(label, value)
	assert.NoErrorf(t, err, "creating a package with label '%s' and value '%s' should not raise an error", label, value)
	return pkg
}

func TestNewPackageFromImage(t *testing.T) {
	// with tag
	pkg, err := NewPackageFromImage("whalebrew/foo:bar", types.ImageInspect{})
	assert.Nil(t, err)
	assert.Equal(t, pkg.Name, "foo")
	assert.Equal(t, pkg.Image, "whalebrew/foo:bar")

	assert.Equal(t, "ws", mustNewTestPkg(t, "io.whalebrew.name", "ws").Name)
	assert.Equal(t, "whalebrew/whalesay", mustNewTestPkg(t, "io.whalebrew.name", "ws").Image)
	assert.Equal(t, []string{"SOME_CONFIG_OPTION"}, mustNewTestPkg(t, "io.whalebrew.config.environment", `["SOME_CONFIG_OPTION"]`).Environment)
	assert.Equal(t, []string{"/somesource:/somedest"}, mustNewTestPkg(t, "io.whalebrew.config.volumes", `["/somesource:/somedest"]`).Volumes)
	assert.Equal(t, []string{"8100:8100"}, mustNewTestPkg(t, "io.whalebrew.config.ports", `["8100:8100"]`).Ports)
	assert.Equal(t, []string{"host"}, mustNewTestPkg(t, "io.whalebrew.config.networks", `["host"]`).Networks)

	assert.True(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "mount").MountMissingVolumes)
	assert.False(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "mount").SkipMissingVolumes)

	assert.False(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "skip").MountMissingVolumes)
	assert.True(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "skip").SkipMissingVolumes)

	assert.False(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "error").MountMissingVolumes)
	assert.False(t, mustNewTestPkg(t, "io.whalebrew.config.missing_volumes", "error").SkipMissingVolumes)

	assert.False(t, mustNewTestPkg(t, "any", "ws").MountMissingVolumes)
	assert.False(t, mustNewTestPkg(t, "any", "other").SkipMissingVolumes)

	_, err = newTestPkg("io.whalebrew.config.missing_volumes", "some-other")
	assert.Error(t, err)
}

func TestPreinstallMessage(t *testing.T) {
	pkg := &Package{}
	assert.Equal(t, pkg.PreinstallMessage(), "")

	pkg = &Package{
		Environment: []string{"AWS_ACCESS_KEY"},
		Ports: []string{
			"80:80",
			"81:81:udp",
		},
		Volumes: []string{
			"/etc/passwd:/passwdtosteal",
			"/etc/readonly:/readonly:ro",
		},
	}
	assert.Equal(t, pkg.PreinstallMessage(),
		"This package needs additional access to your system. It wants to:\n"+
			"\n"+
			"* Read the environment variable AWS_ACCESS_KEY\n"+
			"* Listen on TCP port 80\n"+
			"* Listen on UDP port 81\n"+
			"* Read and write to the file or directory \"/etc/passwd\"\n"+
			"* Read the file or directory \"/etc/readonly\"\n",
	)
}
