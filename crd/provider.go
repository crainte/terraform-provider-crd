package crd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"config_context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"crd_yaml": resourceCRD(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
		ConfigureFunc:  providerConfigure,
	}
}

type KubeClientSet struct {
	Dynamic dynamic.Interface
	Main    *kubernetes.Clientset
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var cfg *restclient.Config
	var err error

	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}

	ctx, ctxOk := d.GetOk("config_context")
	if ctxOk {
		overrides.CurrentContext = ctx.(string)
		overrides.Context = clientcmdapi.Context{}
		log.Printf("[DEBUG] Using custom current context: %q", overrides.CurrentContext)
	}

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	cfg, err = cc.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Failed to load config: %s", err)
	}

	log.Printf("[INFO] Successfully loaded config file")

	// Overriding with static configuration
	cfg.UserAgent = fmt.Sprintf("HashiCorp/1.0 Terraform/%s", terraform.VersionString())

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to configure: %s", err)
	}

	kub, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to configure: %s", err)
	}

	return &KubeClientSet{dyn, kub}, nil
}
