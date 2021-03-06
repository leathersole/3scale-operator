package component

import "fmt"

type AmpImagesOptions struct {
	appLabel                    string
	ampRelease                  string
	apicastImage                string
	backendImage                string
	routerImage                 string
	systemImage                 string
	zyncImage                   string
	ZyncDatabasePostgreSQLImage string
	systemMemcachedImage        string
	systemMySQLImage            string
	insecureImportPolicy        bool
}

type AmpImagesOptionsBuilder struct {
	options AmpImagesOptions
}

func (ampImages *AmpImagesOptionsBuilder) AppLabel(appLabel string) {
	ampImages.options.appLabel = appLabel
}

func (ampImages *AmpImagesOptionsBuilder) AMPRelease(ampRelease string) {
	ampImages.options.ampRelease = ampRelease
}

func (ampImages *AmpImagesOptionsBuilder) ApicastImage(apicastImage string) {
	ampImages.options.apicastImage = apicastImage
}

func (ampImages *AmpImagesOptionsBuilder) BackendImage(backendImage string) {
	ampImages.options.backendImage = backendImage
}

func (ampImages *AmpImagesOptionsBuilder) SystemImage(systemImage string) {
	ampImages.options.systemImage = systemImage
}

func (ampImages *AmpImagesOptionsBuilder) ZyncImage(zyncImage string) {
	ampImages.options.zyncImage = zyncImage
}

func (ampImages *AmpImagesOptionsBuilder) ZyncDatabasePostgreSQLImage(zyncDatabaseImage string) {
	ampImages.options.ZyncDatabasePostgreSQLImage = zyncDatabaseImage
}

func (ampImages *AmpImagesOptionsBuilder) SystemMemcachedImage(image string) {
	ampImages.options.systemMemcachedImage = image
}

func (ampImages *AmpImagesOptionsBuilder) InsecureImportPolicy(insecureImportPolicy bool) {
	ampImages.options.insecureImportPolicy = insecureImportPolicy
}

func (ampImages *AmpImagesOptionsBuilder) Build() (*AmpImagesOptions, error) {
	if ampImages.options.appLabel == "" {
		return nil, fmt.Errorf("no AppLabel has been provided")
	}
	if ampImages.options.ampRelease == "" {
		return nil, fmt.Errorf("no AMP release has been provided")
	}
	if ampImages.options.apicastImage == "" {
		return nil, fmt.Errorf("no Apicast image has been provided")
	}
	if ampImages.options.backendImage == "" {
		return nil, fmt.Errorf("no Backend image has been provided")
	}
	if ampImages.options.systemImage == "" {
		return nil, fmt.Errorf("no System image has been provided")
	}
	if ampImages.options.zyncImage == "" {
		return nil, fmt.Errorf("no Zync image has been provided")
	}
	if ampImages.options.ZyncDatabasePostgreSQLImage == "" {
		return nil, fmt.Errorf("no Zync database PostgreSQL image has been provided")
	}
	if ampImages.options.systemMemcachedImage == "" {
		return nil, fmt.Errorf("no System Memcached image has been provided")
	}

	return &ampImages.options, nil
}
