package k8s

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const GameNamespace = "rpg-demo"

var DungeonGVR = schema.GroupVersionResource{
	Group: "game.k8s.example", Version: "v1alpha1", Resource: "dungeons",
}

type Client struct {
	Dynamic dynamic.Interface
}

func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("backend must run inside the local Kubernetes cluster: %w", err)
	}
	config.Timeout = 10 * time.Second
	config.UserAgent = "krombat-local-backend"
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{Dynamic: dyn}, nil
}
