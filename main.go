package main

import (
	"flag"
	"fmt"
	internalapi "k8s.io/cri-api/pkg/apis"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
	"os"
	"time"
)

// InternalAPIClient is the CRI client.
type InternalAPIClient struct {
	CRIRuntimeClient internalapi.RuntimeService
	CRIImageClient   internalapi.ImageManagerService
}

type ContextType struct {
	// CRI client configurations.
	ImageServiceAddr      string
	ImageServiceTimeout   time.Duration
	RuntimeServiceAddr    string
	RuntimeServiceTimeout time.Duration
}

var Context ContextType

func RegisterFlags() {
	flag.StringVar(&Context.ImageServiceAddr, "image-service-address", "unix:///var/run/crio/crio.sock", "Image service socket for client to connect.")
	flag.DurationVar(&Context.ImageServiceTimeout, "image-service-timeout", 300*time.Second, "Timeout when trying to connect to image service.")
	flag.StringVar(&Context.RuntimeServiceAddr, "runtime-service-address", "unix:///var/run/crio/crio.sock", "Runtime service socket for client to connect..")
	flag.DurationVar(&Context.RuntimeServiceTimeout, "runtime-service-timeout", 300*time.Second, "Timeout when trying to connect to a runtime service.")
	flag.Parse()
}

func LoadCRIClient() (*InternalAPIClient, error) {
	rService, err := remote.NewRemoteRuntimeService(Context.RuntimeServiceAddr, Context.RuntimeServiceTimeout)
	if err != nil {
		return nil, err
	}

	iService, err := remote.NewRemoteImageService(Context.ImageServiceAddr, Context.ImageServiceTimeout)
	if err != nil {
		return nil, err
	}

	return &InternalAPIClient{
		CRIRuntimeClient: rService,
		CRIImageClient:   iService,
	}, nil
}

func ListImage(c internalapi.ImageManagerService, filter *runtimeapi.ImageFilter) []*runtimeapi.Image {
	images, err := c.ListImages(filter)
	if err != nil {
		fmt.Printf("Error listing images: %v", err)
		return []*runtimeapi.Image{}
	}
	return images
}

func ImageStatus(c internalapi.ImageManagerService, imageName string) *runtimeapi.Image {
	imageSpec := &runtimeapi.ImageSpec{
		Image: imageName,
	}
	status, err := c.ImageStatus(imageSpec)
	if err != nil {
		fmt.Printf("Error getting status for image '%s': %v", imageName, err)
		return nil
	}
	return status
}

func main() {
	RegisterFlags()

	c, err := LoadCRIClient()
	if err != nil {
		fmt.Printf("Error loading client: %v", err)
		os.Exit(1)
	}

	images := ListImage(c.CRIImageClient, nil)
	if len(images) == 0 {
		fmt.Printf("No images found")
		os.Exit(0)
	}
	fmt.Printf("\nListing %d images...\n", len(images))
	for _, i := range images {
		fmt.Printf("%s\n", i.String())
	}

	fmt.Printf("\nImageFsInfo():\n")
	fsInfo, err := c.CRIImageClient.ImageFsInfo()
	if err != nil {
		fmt.Printf("Error fetching filesystem usage: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", fsInfo)
}
