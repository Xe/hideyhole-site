// Code generated by protoc-gen-go.
// source: google.golang.org/genproto/googleapis/api/serviceconfig/monitoring.proto
// DO NOT EDIT!

package google_api // import "google.golang.org/genproto/googleapis/api/serviceconfig"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Monitoring configuration of the service.
//
// The example below shows how to configure monitored resources and metrics
// for monitoring. In the example, a monitored resource and two metrics are
// defined. The `library.googleapis.com/book/returned_count` metric is sent
// to both producer and consumer projects, whereas the
// `library.googleapis.com/book/overdue_count` metric is only sent to the
// consumer project.
//
//     monitored_resources:
//     - type: library.googleapis.com/branch
//       labels:
//       - key: /city
//         description: The city where the library branch is located in.
//       - key: /name
//         description: The name of the branch.
//     metrics:
//     - name: library.googleapis.com/book/returned_count
//       metric_kind: DELTA
//       value_type: INT64
//       labels:
//       - key: /customer_id
//     - name: library.googleapis.com/book/overdue_count
//       metric_kind: GAUGE
//       value_type: INT64
//       labels:
//       - key: /customer_id
//     monitoring:
//       producer_destinations:
//       - monitored_resource: library.googleapis.com/branch
//         metrics:
//         - library.googleapis.com/book/returned_count
//       consumer_destinations:
//       - monitored_resource: library.googleapis.com/branch
//         metrics:
//         - library.googleapis.com/book/returned_count
//         - library.googleapis.com/book/overdue_count
//
type Monitoring struct {
	// Monitoring configurations for sending metrics to the producer project.
	// There can be multiple producer destinations, each one must have a
	// different monitored resource type. A metric can be used in at most
	// one producer destination.
	ProducerDestinations []*Monitoring_MonitoringDestination `protobuf:"bytes,1,rep,name=producer_destinations,json=producerDestinations" json:"producer_destinations,omitempty"`
	// Monitoring configurations for sending metrics to the consumer project.
	// There can be multiple consumer destinations, each one must have a
	// different monitored resource type. A metric can be used in at most
	// one consumer destination.
	ConsumerDestinations []*Monitoring_MonitoringDestination `protobuf:"bytes,2,rep,name=consumer_destinations,json=consumerDestinations" json:"consumer_destinations,omitempty"`
}

func (m *Monitoring) Reset()                    { *m = Monitoring{} }
func (m *Monitoring) String() string            { return proto.CompactTextString(m) }
func (*Monitoring) ProtoMessage()               {}
func (*Monitoring) Descriptor() ([]byte, []int) { return fileDescriptor11, []int{0} }

func (m *Monitoring) GetProducerDestinations() []*Monitoring_MonitoringDestination {
	if m != nil {
		return m.ProducerDestinations
	}
	return nil
}

func (m *Monitoring) GetConsumerDestinations() []*Monitoring_MonitoringDestination {
	if m != nil {
		return m.ConsumerDestinations
	}
	return nil
}

// Configuration of a specific monitoring destination (the producer project
// or the consumer project).
type Monitoring_MonitoringDestination struct {
	// The monitored resource type. The type must be defined in
	// [Service.monitored_resources][google.api.Service.monitored_resources] section.
	MonitoredResource string `protobuf:"bytes,1,opt,name=monitored_resource,json=monitoredResource" json:"monitored_resource,omitempty"`
	// Names of the metrics to report to this monitoring destination.
	// Each name must be defined in [Service.metrics][google.api.Service.metrics] section.
	Metrics []string `protobuf:"bytes,2,rep,name=metrics" json:"metrics,omitempty"`
}

func (m *Monitoring_MonitoringDestination) Reset()         { *m = Monitoring_MonitoringDestination{} }
func (m *Monitoring_MonitoringDestination) String() string { return proto.CompactTextString(m) }
func (*Monitoring_MonitoringDestination) ProtoMessage()    {}
func (*Monitoring_MonitoringDestination) Descriptor() ([]byte, []int) {
	return fileDescriptor11, []int{0, 0}
}

func init() {
	proto.RegisterType((*Monitoring)(nil), "google.api.Monitoring")
	proto.RegisterType((*Monitoring_MonitoringDestination)(nil), "google.api.Monitoring.MonitoringDestination")
}

func init() {
	proto.RegisterFile("google.golang.org/genproto/googleapis/api/serviceconfig/monitoring.proto", fileDescriptor11)
}

var fileDescriptor11 = []byte{
	// 255 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x90, 0x41, 0x4b, 0xc3, 0x40,
	0x10, 0x85, 0x49, 0x05, 0xa5, 0x2b, 0x28, 0x2e, 0x16, 0x4a, 0x4f, 0x45, 0x2f, 0x3d, 0xe8, 0x2e,
	0xe8, 0x3f, 0x28, 0x1e, 0xf4, 0x20, 0x94, 0xfc, 0x81, 0xba, 0x6e, 0xc6, 0x65, 0xa0, 0x99, 0x59,
	0x66, 0x37, 0xfe, 0x32, 0x7f, 0xa0, 0xb4, 0x49, 0x9a, 0x20, 0x9e, 0x7a, 0x4b, 0xf6, 0xbd, 0x79,
	0xdf, 0xe3, 0xa9, 0xd7, 0xc0, 0x1c, 0x76, 0x60, 0x02, 0xef, 0x1c, 0x05, 0xc3, 0x12, 0x6c, 0x00,
	0x8a, 0xc2, 0x99, 0x6d, 0x2b, 0xb9, 0x88, 0xc9, 0xba, 0x88, 0x36, 0x81, 0x7c, 0xa3, 0x07, 0xcf,
	0xf4, 0x85, 0xc1, 0xd6, 0x4c, 0x98, 0x59, 0x90, 0x82, 0x39, 0xb8, 0xb5, 0xea, 0x92, 0x5c, 0xc4,
	0xc5, 0xdb, 0xa9, 0xa9, 0x8e, 0x88, 0xb3, 0xcb, 0xc8, 0x94, 0xda, 0xd8, 0xbb, 0x9f, 0x89, 0x52,
	0xef, 0x47, 0x96, 0x76, 0x6a, 0x16, 0x85, 0xab, 0xc6, 0x83, 0x6c, 0x2b, 0x48, 0x19, 0xa9, 0x75,
	0xcf, 0x8b, 0xe5, 0xd9, 0xea, 0xf2, 0xe9, 0xc1, 0x0c, 0x2d, 0xcc, 0x70, 0x36, 0xfa, 0x7c, 0x19,
	0x8e, 0xca, 0xdb, 0x3e, 0x6a, 0xf4, 0x98, 0xf6, 0x08, 0xcf, 0x94, 0x9a, 0xfa, 0x2f, 0x62, 0x72,
	0x0a, 0xa2, 0x8f, 0x1a, 0x23, 0x16, 0x1f, 0x6a, 0xf6, 0xaf, 0x5d, 0x3f, 0x2a, 0xdd, 0x0d, 0x0b,
	0xd5, 0x56, 0x20, 0x71, 0x23, 0x1e, 0xe6, 0xc5, 0xb2, 0x58, 0x4d, 0xcb, 0x9b, 0xa3, 0x52, 0x76,
	0x82, 0x9e, 0xab, 0x8b, 0x1a, 0xb2, 0xa0, 0x6f, 0xcb, 0x4d, 0xcb, 0xfe, 0x77, 0x7d, 0xaf, 0xae,
	0x3c, 0xd7, 0xa3, 0xaa, 0xeb, 0xeb, 0x81, 0xb8, 0xd9, 0x2f, 0xbb, 0x29, 0x3e, 0xcf, 0x0f, 0x13,
	0x3f, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff, 0xf5, 0x1e, 0xd7, 0x88, 0x05, 0x02, 0x00, 0x00,
}
