package protor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/rubyist/circuitbreaker"
)

type Protor struct {
	PO ProtorOption
	cb *circuit.Breaker
}

type ProtorOption struct {
	Kind    string
	Address string
}

func NewProtor(PO ProtorOption) *Protor {
	cb := circuit.NewRateBreaker(0.1, 100)
	return &Protor{
		PO: PO,
		cb: cb,
	}
}

type MetricData struct {
	Start  time.Time
	Action string
	Status string
}

type ProtorData struct {
	Name       string            `json:"name"`
	Kind       string            `json:"kind"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	Additional []float64         `json:"additional"`
}

func Metric(ctx context.Context, m *MetricData) *ProtorData {

	sample := &ProtorData{
		Name:       "service_latency_seconds",
		Value:      time.Since(m.Start).Seconds(),
		Kind:       "h",
		Additional: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		Labels: map[string]string{
			"action": m.Action,
			"status": m.Status,
			"job":    "attache",
		},
	}
	return sample
}

func (p *Protor) Work(ctx context.Context, sample *ProtorData) error {
	if p.cb.Ready() {
		conn, err := net.Dial(p.PO.Kind, p.PO.Address)
		for err != nil {
			p.cb.Fail()
			return err
		}
		p.cb.Success()
		defer conn.Close()
		if p.Valid(ctx, sample) {
			fmt.Fprintf(conn, p.Encode(ctx, sample))
		}
	} else {
		return errors.New("Circuit Breaker is in Trip State")
	}
	return nil
}

func (p *Protor) Decode(ctx context.Context, r io.Reader) (samples []*ProtorData, err error) {
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&samples)
	return
}

func (p *Protor) Valid(ctx context.Context, s *ProtorData) bool {
	b := (s.Kind == "c" || s.Kind == "g" || s.Kind == "hl" || s.Kind == "h") && len(s.Name) > 0
	return b
}

func (p *Protor) Encode(ctx context.Context, pd *ProtorData) string {
	var buffer bytes.Buffer
	buffer.WriteString(namePart(pd))
	buffer.WriteString(kindPart(pd))

	if pd.Kind == "h" || pd.Kind == "hl" {
		buffer.WriteString(additionalPart(pd))
	}

	buffer.WriteString(labelPart(pd))
	buffer.WriteString(valuePart(pd))

	return buffer.String()
}

func namePart(pd *ProtorData) string {
	return fmt.Sprintf("%s|", pd.Name)
}

func kindPart(pd *ProtorData) string {
	return fmt.Sprintf("%s|", pd.Kind)
}

func valuePart(pd *ProtorData) string {
	return fmt.Sprintf("%.6f", pd.Value)
}

func additionalPart(pd *ProtorData) string {
	additional := pd.Additional
	if len(additional) == 0 {
		if pd.Kind == "h" {
			additional = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
		} else {
			return ""
		}
	}

	var buffer bytes.Buffer
	for i, v := range additional {
		if i != 0 {
			buffer.WriteString(";")
		}
		buffer.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	}
	buffer.WriteString("|")
	return buffer.String()
}

func labelPart(pd *ProtorData) string {
	if len(pd.Labels) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	i := 0
	for k, v := range pd.Labels {
		buffer.WriteString(k)
		buffer.WriteString("=")
		buffer.WriteString(v)
		i++

		if i != len(pd.Labels) {
			buffer.WriteString(";")
		}
	}
	buffer.WriteString("|")
	return buffer.String()
}
