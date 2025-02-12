package chart

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ketchv1 "github.com/theketchio/ketch/internal/api/v1beta1"
)

func TestNewIngress(t *testing.T) {
	tests := []struct {
		name          string
		cnames        ketchv1.CnameList
		clusterIssuer string
		expected      *ingress
		expectedError error
	}{
		{
			name: "happy",
			cnames: ketchv1.CnameList{
				{Name: "a.name"},
				{Name: "b.name", Secure: true},
				{Name: "c.name", Secure: true, SecretName: "c-ssl"},
			},
			clusterIssuer: "test-cluster-issuer",
			expected: &ingress{
				Http: []string{"a.name"},
				Https: []httpsEndpoint{
					{Cname: "b.name", SecretName: "my-app-cname-b-name", UniqueName: "my-app-https-b-name", ManagedBy: certManager},
					{Cname: "c.name", SecretName: "c-ssl", UniqueName: "my-app-https-c-name", ManagedBy: user},
				},
			},
		},
		{
			name: "happy - no https, no cluster issuer",
			cnames: ketchv1.CnameList{
				{
					Name: "a.name",
				},
				{
					Name: "b.name",
				},
			},
			expected: &ingress{
				Http: []string{"a.name", "b.name"},
			},
		},
		{
			name: "sad - no cluster issuer",
			cnames: ketchv1.CnameList{
				{
					Name:   "a.name",
					Secure: true,
				},
			},
			expectedError: errors.New("secure cnames require a framework.Ingress.ClusterIssuer to be specified"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := ketchv1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-app",
				},
				Spec: ketchv1.AppSpec{
					Ingress: ketchv1.IngressSpec{
						Cnames: tt.cnames,
					},
				},
			}
			framework := ketchv1.Framework{
				Spec: ketchv1.FrameworkSpec{
					IngressController: ketchv1.IngressControllerSpec{
						ClusterIssuer: tt.clusterIssuer,
					},
				},
			}
			issuer, err := newIngress(app, framework)
			if tt.expectedError != nil {
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.Equal(t, tt.expected, issuer)
			}
		})
	}
}
