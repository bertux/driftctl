package terraform

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/hashicorp/go-getter"

	"github.com/sirupsen/logrus"
)

type ProviderDownloaderInterface interface {
	Download(url, path string) error
	GetProviderUrl(name, version string) string
}

type ProviderDownloader struct {
	httpclient *http.Client
	unzip      getter.ZipDecompressor
	context    context.Context
}

func NewProviderDownloader() *ProviderDownloader {
	return &ProviderDownloader{
		httpclient: http.DefaultClient,
		unzip:      getter.ZipDecompressor{},
		context:    context.Background(),
	}
}

func (p *ProviderDownloader) GetProviderUrl(name, version string) string {
	return fmt.Sprintf(
		"https://releases.hashicorp.com/terraform-provider-%s/%s/terraform-provider-%s_%s_%s_%s.zip",
		name,
		version,
		name,
		version,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func (p *ProviderDownloader) Download(url, path string) error {
	logrus.WithFields(logrus.Fields{
		"url":  url,
		"path": path,
	}).Debug("Downloading provider")

	req, err := http.NewRequestWithContext(p.context, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := p.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unsuccessful request to %s: %s", url, resp.Status)
	}
	f, err := ioutil.TempFile("", "terraform-provider")
	if err != nil {
		return fmt.Errorf("failed to open temporary file to download from %s", url)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	n, err := getter.Copy(p.context, f, resp.Body)
	if err == nil && n < resp.ContentLength {
		err = fmt.Errorf("incorrect response size: expected %d bytes, but got %d bytes", resp.ContentLength, n)
	}
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"src": f.Name(),
		"dst": path,
	}).Debug("Decompressing archive")
	err = p.unzip.Decompress(path, f.Name(), true, 0)
	if err != nil {
		return err
	}
	return nil
}
