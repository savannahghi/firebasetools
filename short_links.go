package firebasetools

import (
	"context"
	"fmt"

	"github.com/savannahghi/serverutils"
	"google.golang.org/api/firebasedynamiclinks/v1"
)

// ShortenLink shortens an FDL link
func ShortenLink(ctx context.Context, longLink string) (string, error) {
	fdlService, err := firebasedynamiclinks.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to initialize Firebase Dynamic Links service: %w", err)
	}

	dynamicLinkDomain, err := serverutils.GetEnvVar(FDLDomainEnvironmentVariableName)
	if err != nil {
		return "", fmt.Errorf("environment variable %s missing", FDLDomainEnvironmentVariableName)
	}

	linkRequest := &firebasedynamiclinks.CreateShortDynamicLinkRequest{
		DynamicLinkInfo: &firebasedynamiclinks.DynamicLinkInfo{
			DomainUriPrefix: dynamicLinkDomain,
			Link:            longLink,
		},
	}
	linkReq := fdlService.ShortLinks.Create(linkRequest)
	linkResp, err := linkReq.Do()
	if err != nil {
		return "", fmt.Errorf("unable to shorten link: %w", err)
	}
	return linkResp.ShortLink, nil
}
