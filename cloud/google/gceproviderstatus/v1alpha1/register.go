/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"fmt"
	"bytes"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/cloud/google/gceproviderstatus"
)

type GCEProviderStatusCodec struct {
	encoder runtime.Encoder
	decoder runtime.Decoder
}

var (
	SchemeBuilder      runtime.SchemeBuilder
	AddToScheme        = SchemeBuilder.AddToScheme
	localSchemeBuilder = &SchemeBuilder
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

const GroupName = "gceproviderstatus"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&GCEClusterProviderStatus{},
	)
	return nil
}

func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := gceproviderstatus.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func NewCodec() (*GCEProviderStatusCodec, error) {
	scheme, err := NewScheme()
	if err != nil {
		return nil, err
	}
	codecFactory := serializer.NewCodecFactory(scheme)
	encoder, err := newEncoder(&codecFactory)
	if err != nil {
		return nil, err
	}
	codec := GCEProviderStatusCodec{
		encoder: encoder,
		decoder: codecFactory.UniversalDecoder(SchemeGroupVersion),
	}
	return &codec, nil
}

func (codec *GCEProviderStatusCodec) DecodeFromProviderStatus(providerStatus clusterv1.ProviderStatus, out runtime.Object) (error) {
	_, _, err := codec.decoder.Decode(providerStatus.Value.Raw, nil, out)
	if err != nil {
		return fmt.Errorf("decoding failure: %v", err)
	}
	return nil
}

func (codec *GCEProviderStatusCodec) EncodeToProviderStatus(in runtime.Object) (*clusterv1.ProviderStatus, error) {
	var buf bytes.Buffer
	if err := codec.encoder.Encode(in, &buf); err != nil {
		return nil, fmt.Errorf("encoding failed: %v", err)
	}
	return &clusterv1.ProviderStatus{
		Value: &runtime.RawExtension{Raw: buf.Bytes()},
	}, nil
}

func newEncoder(codecFactory *serializer.CodecFactory) (runtime.Encoder, error) {
	serializerInfos := codecFactory.SupportedMediaTypes()
	if len(serializerInfos) == 0 {
		return nil, fmt.Errorf("unable to find any serlializers")
	}
	encoder := codecFactory.EncoderForVersion(serializerInfos[0].Serializer, SchemeGroupVersion)
	return encoder, nil
}
