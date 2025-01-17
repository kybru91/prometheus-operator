// Copyright 2021 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operator

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/prometheus-operator/prometheus-operator/pkg/assets"
	"github.com/prometheus-operator/prometheus-operator/pkg/k8sutil"
)

// MaxSecretDataSizeBytes is the maximum data size that a single secret shard
// may use. This is lower than v1.MaxSecretSize in order to reserve space for
// metadata and the rest of the secret k8s object.
const MaxSecretDataSizeBytes = v1.MaxSecretSize - 50_000

// ShardedSecret can shard Secret data across multiple k8s Secrets.
// This is used to circumvent the size limitation of k8s Secrets.
type ShardedSecret struct {
	template     *v1.Secret
	data         map[string][]byte
	secretShards []*v1.Secret
}

// NewShardedSecret takes a v1.Secret object as template and returns a new ShardedSecret.
// The template's name will be used as the prefix for the concrete secrets.
func NewShardedSecret(template *v1.Secret) *ShardedSecret {
	return &ShardedSecret{
		template: template,
		data:     make(map[string][]byte),
	}
}

type Byter interface {
	Bytes() []byte
}

// Append adds a new key + data pair.
// If the key already exists, data gets overwritten.
func (s *ShardedSecret) Append(k fmt.Stringer, v Byter) {
	s.data[k.String()] = v.Bytes()
}

// UpdateSecrets updates the concrete Secrets from the stored data.
func (s *ShardedSecret) UpdateSecrets(ctx context.Context, sClient corev1.SecretInterface) error {
	secrets := s.shard()

	for _, secret := range secrets {
		err := k8sutil.CreateOrUpdateSecret(ctx, sClient, secret)
		if err != nil {
			return fmt.Errorf("failed to create secret %q: %w", secret.Name, err)
		}
	}

	return s.cleanupExcessSecretShards(ctx, sClient, len(secrets)-1)
}

// shard does the in-memory sharding of the secret data.
func (s *ShardedSecret) shard() []*v1.Secret {
	s.secretShards = []*v1.Secret{}

	// Ensure that we always iterate over the keys in the same order.
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	currentIndex := 0
	secretSize := 0
	currentSecret := s.newSecretAt(currentIndex)

	for _, key := range keys {
		v := s.data[key]
		vSize := len(key) + len(v)
		if secretSize+vSize > MaxSecretDataSizeBytes {
			s.secretShards = append(s.secretShards, currentSecret)
			currentIndex++
			secretSize = 0
			currentSecret = s.newSecretAt(currentIndex)
		}

		secretSize += vSize
		currentSecret.Data[key] = v
	}
	s.secretShards = append(s.secretShards, currentSecret)

	return s.secretShards
}

// newSecretAt creates a new Kubernetes object at the given shard index.
func (s *ShardedSecret) newSecretAt(index int) *v1.Secret {
	newShardSecret := s.template.DeepCopy()
	newShardSecret.Name = s.secretNameAt(index)
	newShardSecret.Data = make(map[string][]byte)

	return newShardSecret
}

// cleanupExcessSecretShards removes excess secret shards that are no longer in use.
// It also tries to remove a non-sharded secret that exactly matches the name
// prefix in order to make sure that operator version upgrades run smoothly.
func (s *ShardedSecret) cleanupExcessSecretShards(ctx context.Context, sClient corev1.SecretInterface, lastSecretIndex int) error {
	for i := lastSecretIndex + 1; ; i++ {
		secretName := s.secretNameAt(i)
		err := sClient.Delete(ctx, secretName, metav1.DeleteOptions{})
		if apierrors.IsNotFound(err) {
			// we reached the end of existing secrets
			break
		}

		if err != nil {
			return fmt.Errorf("failed to delete secret %q: %w", secretName, err)
		}
	}

	return nil
}

func (s *ShardedSecret) secretNameAt(index int) string {
	return fmt.Sprintf("%s-%d", s.template.Name, index)
}

// SecretNames returns the names of the concrete secrets.
// It must be called after UpdateSecrets().
func (s *ShardedSecret) SecretNames() []string {
	var names []string
	for i := 0; i < len(s.secretShards); i++ {
		names = append(names, s.secretNameAt(i))
	}

	return names
}

func ReconcileShardedSecretForTLSAssets(ctx context.Context, store *assets.Store, client kubernetes.Interface, template *v1.Secret) (*ShardedSecret, error) {
	shardedSecret := NewShardedSecret(template)

	for k, v := range store.TLSAssets {
		shardedSecret.Append(k, v)
	}

	if err := shardedSecret.UpdateSecrets(ctx, client.CoreV1().Secrets(template.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to update the TLS secrets: %w", err)
	}

	return shardedSecret, nil
}
