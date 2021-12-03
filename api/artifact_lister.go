package api

type ArtifactLister interface {
	ListBuildArtifacts(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error)
}

type defaultArtifactLister struct {
	apiClient BitriseAPIClient
}

func NewArtifactLister(client BitriseAPIClient) ArtifactLister {
	return defaultArtifactLister{
		apiClient: client,
	}
}

func (lister defaultArtifactLister) ListBuildArtifacts(appSlug string, buildSlugs []string) ([]ArtifactResponseItemModel, error) {
	var artifacts []ArtifactResponseItemModel

	// wg := &sync.WaitGroup{}
	// wg.Add(len(buildSlugs))
	// for i, buildSlug := range buildSlugs {

	// }
	// wg.Wait()

	return artifacts, nil
}
