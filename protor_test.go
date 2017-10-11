package protor_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	pt "github.com/bukalapak/protor"
)

type ProtorSuite struct {
	suite.Suite
	protor *pt.Protor
}

func TestProtorSuite(t *testing.T) {
	suite.Run(t, &ProtorSuite{})
}

func (ps *ProtorSuite) SetupSuite() {
	ps.protor = pt.DefaultProtor()
	ps.protor.Option.Kind = "tcp"
	ps.protor.Option.Address = "localhost:3000"
}

func (ps *ProtorSuite) TestMetric() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	start := time.Now()
	MD := &pt.MetricData{
		Start:  start,
		Action: "ingest",
		Status: "ok",
	}
	data := pt.Metric(ctx, MD)
	assert.NotNil(ps.T(), data, "")
}

func (ps *ProtorSuite) TestWork() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	start := time.Now()
	MD := &pt.MetricData{
		Start:  start,
		Action: "ingest",
		Status: "ok",
	}
	///
	go func() {
		conn, err := net.Dial("tcp", ":3000")
		assert.Nil(ps.T(), err, "error connecting")
		defer conn.Close()
	}()
	l, err := net.Listen("tcp", ":3000")
	assert.Nil(ps.T(), err, "can't open server")

	defer l.Close()

	err = ps.protor.Work(ctx, pt.Metric(ctx, MD))
	i := 0
	for i < 2 {
		i = i + 1
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, err := ioutil.ReadAll(conn)
		assert.Nil(ps.T(), err, "error reading")

		fmt.Println(string(buf[:]))
	}
}

func (ps *ProtorSuite) TestDecode() {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	xamplejson, _ := ioutil.ReadFile("testdata.in")
	r := bytes.NewReader(xamplejson)
	data, _ := ps.protor.Decode(ctx, r)

	assert.NotNil(ps.T(), data, "Protor Decode ERROR")
}

func (ps *ProtorSuite) TestValid() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	xamplejson, _ := ioutil.ReadFile("jsonxample")
	r := bytes.NewReader(xamplejson)
	samples, _ := ps.protor.Decode(ctx, r)
	for _, sample := range samples {
		data := ps.protor.Encode(ctx, sample)
		assert.NotNil(ps.T(), data, "error")
	}
}
