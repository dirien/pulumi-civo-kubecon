package main

import (
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		firewall, err := civo.NewFirewall(ctx, "civo-firewall", &civo.FirewallArgs{
			Name:               pulumi.String("myFirstFirewall"),
			Region:             pulumi.StringPtr("KUBECON"),
			CreateDefaultRules: pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}
		cluster, err := civo.NewKubernetesCluster(ctx, "civo-k3s-cluster-kubecon", &civo.KubernetesClusterArgs{
			Name: pulumi.StringPtr("civo-k3s-cluster-kubecon"),
			Pools: civo.KubernetesClusterPoolsArgs{
				Size:      pulumi.String("g4s.kube.medium"),
				NodeCount: pulumi.Int(1),
			},
			Region:     pulumi.StringPtr("KUBECON"),
			FirewallId: firewall.ID(),
		})
		if err != nil {
			return err
		}

		provider, err := kubernetes.NewProvider(ctx, "kubernetes", &kubernetes.ProviderArgs{
			Kubeconfig: cluster.Kubeconfig,
		})
		if err != nil {
			return err
		}
		_, err = helm.NewRelease(ctx, "minecraft", &helm.ReleaseArgs{
			Name:            pulumi.String("minecraft"),
			Chart:           pulumi.String("minecraft"),
			Version:         pulumi.String("4.0.0"),
			Namespace:       pulumi.String("minecraft"),
			CreateNamespace: pulumi.Bool(true),
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://itzg.github.io/minecraft-server-charts/"),
			},
			ValueYamlFiles: pulumi.AssetOrArchiveArray{
				pulumi.NewFileAsset("values/minecraft.yaml"),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		ctx.Export("name", cluster.Name)
		ctx.Export("kubeconfig", cluster.Kubeconfig)
		return nil
	})
}
