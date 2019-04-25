package v1alpha1

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for PVC
type Builder struct {
	pvc  *PVC
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pvc: &PVC{object: &corev1.PersistentVolumeClaim{}}}
}

// WithName sets the Name field of PVC with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing PVC name"))
		return b
	}
	b.pvc.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of PVC provided arguments
func (b *Builder) WithNamespace(namespace string) *Builder {
	if len(namespace) == 0 {
		namespace = "default"
	}
	b.pvc.object.Namespace = namespace
	return b
}

// WithAnnotations sets the Annotations field of PVC with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing annotations"))
		return b
	}
	b.pvc.object.Annotations = annotations
	return b
}

// WithLabels sets the Labels field of PVC with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing labels"))
		return b
	}
	b.pvc.object.Labels = labels
	return b
}

// WithStorageClass sets the StorageClass field of PVC with provided arguments
func (b *Builder) WithStorageClass(scName string) *Builder {
	if len(scName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing storageclass name"))
		return b
	}
	b.pvc.object.Spec.StorageClassName = &scName
	return b
}

// WithAccessModes sets the AccessMode field in PVC with provided arguments
func (b *Builder) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *Builder {
	if len(accessMode) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PVC object: missing accessmodes"))
		return b
	}
	b.pvc.object.Spec.AccessModes = accessMode
	return b
}

// WithCapacity sets the Capacity field in PVC with provided arguments
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(b.errs, errors.Errorf("failed to build PVC object: failed to parse capacity {%s}", capacity))
		return b
	}
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	b.pvc.object.Spec.Resources.Requests = resourceList
	return b
}

// Build returns the PVC API instance
func (b *Builder) Build() (*corev1.PersistentVolumeClaim, []error) {
	if len(b.errs) > 0 {
		return nil, b.errs
	}
	return b.pvc.object, nil
}
