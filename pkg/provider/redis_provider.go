/* Redis provider */

package redismetrics

import (
	"github.com/jake-quovo/custom-metrics-apiserver/provider"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/custom_metrics"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

func (p *redisProvider) ListAllMetrics() []provider.CustomMetricInfo {
	return []provider.CustomMetricInfo{
		// these are mostly arbitrary examples
		{
			GroupResource: schema.GroupResource{Group: "", Resource: "pods"},
			Metric:        "packets-per-second",
			Namespaced:    true,
		},
		{
			GroupResource: schema.GroupResource{Group: "", Resource: "services"},
			Metric:        "connections-per-second",
			Namespaced:    true,
		},
		{
			GroupResource: schema.GroupResource{Group: "", Resource: "namespaces"},
			Metric:        "work-queue-length",
			Namespaced:    false,
		},
	}
}

type redisProvider struct {
	client dynamic.Interface
	mapper apimeta.RESTMapper

	// just increment values when they're requested
	values map[provider.CustomMetricInfo]int64
}

// valueFor fetches a value from the fake list and increments it.
func (p *redisProvider) valueFor(info provider.CustomMetricInfo) (int64, error) {
	// normalize the value so that you treat plural resources and singular
	// resources the same (e.g. pods vs pod)
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		return 0, err
	}

	value := p.values[info]
	value += 1
	p.values[info] = value

	return value, nil
}

// metricFor constructs a result for a single metric value.
func (p *redisProvider) metricFor(value int64, name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	// construct a reference referring to the described object
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	return &custom_metrics.MetricValue{
		DescribedObject: objRef,
		MetricName:      info.Metric,
		// you'll want to use the actual timestamp in a real adapter
		Timestamp: metav1.Time{time.Now()},
		Value:     *resource.NewMilliQuantity(value*100, resource.DecimalSI),
	}, nil
}

func (p *redisProvider) GetMetricByName(name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	value, err := p.valueFor(info)
	if err != nil {
		return nil, err
	}
	return p.metricFor(value, name, info)
}
