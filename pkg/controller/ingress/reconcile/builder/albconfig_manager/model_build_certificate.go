package albconfigmanager

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb"
)

func (t *defaultModelBuildTask) buildSecretCertificate(ctx context.Context, ing networking.Ingress, secretName, clusterID string) (*alb.SecretCertificate, error) {
	scSpecDst, err := t.buildSecretCertificateSpec(ctx, ing, secretName, clusterID)
	if err != nil {
		return nil, err
	}
	scResID := fmt.Sprintf("%v", scSpecDst.CertName)
	if sc, exists := t.scByResID[scResID]; exists {
		return sc, nil
	}
	sc := alb.NewSecretCertificate(t.stack, scResID, scSpecDst)
	t.scByResID[scResID] = sc
	return sc, nil
}

func (t *defaultModelBuildTask) buildSecretCertificateSpec(ctx context.Context, ing networking.Ingress, secretName, clusterID string) (alb.SecretCertificateSpec, error) {
	var secret = &corev1.Secret{}
	err := t.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: ing.Namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return alb.SecretCertificateSpec{}, err
	}
	crt := string(secret.Data["tls.crt"])
	key := string(secret.Data["tls.key"])
	digestCert := computeDigest(clusterID, crt, key)

	certName := fmt.Sprintf("%s-%s-%s", secret.Namespace, secret.Name, digestCert)

	sc := alb.SecretCertificateSpec{
		CertName:    certName,
		IsDefault:   false,
		Certificate: crt,
		PrivateKey:  key,
	}
	return sc, nil
}

func computeDigest(args ...string) string {
	data := ""
	for _, a := range args {
		data = data + a
	}
	bData := []byte(data)
	shaSum := sha1.Sum(bData)
	ret := hex.EncodeToString(shaSum[:])[0:6]
	return ret
}
