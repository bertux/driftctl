package terraform

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/stretchr/testify/assert"
)

func TestProviderInstallerInstallDoesNotExist(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()

	expectedSubFolder := fmt.Sprintf("/.driftctl/plugins/%s_%s", runtime.GOOS, runtime.GOARCH)

	config := ProviderConfig{
		Key:     "aws",
		Version: "3.19.0",
		Postfix: "x5",
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}
	mockDownloader.On("Download", config.GetDownloadUrl(), path.Join(fakeTmpHome, expectedSubFolder)).Return(nil)

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		config:     config,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.Install()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(fakeTmpHome, expectedSubFolder, config.GetBinaryName()), providerPath)

}

func TestProviderInstallerInstallWithoutHomeDir(t *testing.T) {

	assert := assert.New(t)

	expectedHomeDir := os.TempDir()
	expectedSubFolder := fmt.Sprintf("/.driftctl/plugins/%s_%s", runtime.GOOS, runtime.GOARCH)
	config := ProviderConfig{
		Key:     "aws",
		Version: "3.19.0",
		Postfix: "x5",
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}
	mockDownloader.On("Download", config.GetDownloadUrl(), path.Join(expectedHomeDir, expectedSubFolder)).Return(nil)

	installer := ProviderInstaller{
		config:     config,
		downloader: &mockDownloader,
	}

	providerPath, err := installer.Install()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(expectedHomeDir, expectedSubFolder, config.GetBinaryName()), providerPath)

}

func TestProviderInstallerInstallAlreadyExist(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()
	expectedSubFolder := fmt.Sprintf("/.driftctl/plugins/%s_%s", runtime.GOOS, runtime.GOARCH)
	err := os.MkdirAll(path.Join(fakeTmpHome, expectedSubFolder), 0755)
	if err != nil {
		t.Error(err)
	}

	config := ProviderConfig{
		Key:     "aws",
		Version: "3.19.0",
		Postfix: "x5",
	}

	_, err = os.Create(path.Join(fakeTmpHome, expectedSubFolder, config.GetBinaryName()))
	if err != nil {
		t.Error(err)
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		config:     config,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.Install()
	mockDownloader.AssertExpectations(t)

	assert.Nil(err)
	assert.Equal(path.Join(fakeTmpHome, expectedSubFolder, config.GetBinaryName()), providerPath)

}

func TestProviderInstallerInstallAlreadyExistButIsDirectory(t *testing.T) {

	assert := assert.New(t)
	fakeTmpHome := t.TempDir()
	expectedSubFolder := fmt.Sprintf("/.driftctl/plugins/%s_%s", runtime.GOOS, runtime.GOARCH)

	config := ProviderConfig{
		Key:     "aws",
		Version: "3.19.0",
		Postfix: "x5",
	}

	invalidDirPath := path.Join(fakeTmpHome, expectedSubFolder, config.GetBinaryName())
	err := os.MkdirAll(invalidDirPath, 0755)
	if err != nil {
		t.Error(err)
	}

	mockDownloader := mocks.ProviderDownloaderInterface{}

	installer := ProviderInstaller{
		downloader: &mockDownloader,
		config:     config,
		homeDir:    fakeTmpHome,
	}

	providerPath, err := installer.Install()
	mockDownloader.AssertExpectations(t)

	assert.Empty(providerPath)
	assert.NotNil(err)
	assert.Equal(
		fmt.Sprintf(
			"found directory instead of provider binary in %s",
			invalidDirPath,
		),
		err.Error(),
	)

}
