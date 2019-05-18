package util

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fagongzi/log"
	"github.com/fagongzi/util/task"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

const contentTypeHeader = "Content-Type"

var (
	client         = &http.Client{}
	defaultTimeout = time.Second * 15
)

// MetricCfg is the metric configuration.
type MetricCfg struct {
	Job          string
	Instance     string
	Address      string
	DurationSync time.Duration
}

// NewMetricCfg returns metric cfg
func NewMetricCfg(job, instance, address string, durationSync time.Duration) *MetricCfg {
	return &MetricCfg{
		Job:          job,
		Instance:     instance,
		Address:      address,
		DurationSync: durationSync,
	}
}

// StartMetricsPush start a push client
func StartMetricsPush(runner *task.Runner, cfg *MetricCfg) {
	if nil == cfg || cfg.DurationSync == 0 || len(cfg.Address) == 0 {
		log.Info("metric: disable prometheus push client")
		return
	}

	*client = *http.DefaultClient
	client.Timeout = defaultTimeout

	log.Info("metric: start prometheus push client")
	runner.RunCancelableTask(func(ctx context.Context) {
		t := time.NewTicker(cfg.DurationSync)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("stop: prometheus push client stopped")
				t.Stop()
				return
			case <-t.C:
				err := doPush(cfg.Job, instanceGroupingKey(cfg.Instance), cfg.Address, prometheus.DefaultGatherer, "PUT")
				if err != nil {
					log.Errorf("metric: could not push metrics to prometheus pushgateway: errors:\n%+v", err)
				}
			}
		}
	})
}

// instanceGroupingKey returns a label map with the only entry
// {instance="<instance>"}. If instance is empty, use hostname instead.
func instanceGroupingKey(instance string) map[string]string {
	if instance == "" {
		var err error
		if instance, err = os.Hostname(); err != nil {
			instance = "unknown"
		}
	}
	return map[string]string{"instance": instance}
}

func doPush(job string, grouping map[string]string, pushURL string, g prometheus.Gatherer, method string) error {
	if !strings.Contains(pushURL, "://") {
		pushURL = "http://" + pushURL
	}
	if strings.HasSuffix(pushURL, "/") {
		pushURL = pushURL[:len(pushURL)-1]
	}

	if strings.Contains(job, "/") {
		return fmt.Errorf("job contains '/': %s", job)
	}
	urlComponents := []string{url.QueryEscape(job)}
	for ln, lv := range grouping {
		if !model.LabelName(ln).IsValid() {
			return fmt.Errorf("grouping label has invalid name: %s", ln)
		}
		if strings.Contains(lv, "/") {
			return fmt.Errorf("value of grouping label %s contains '/': %s", ln, lv)
		}
		urlComponents = append(urlComponents, ln, lv)
	}
	pushURL = fmt.Sprintf("%s/metrics/job/%s", pushURL, strings.Join(urlComponents, "/"))

	mfs, err := g.Gather()
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtProtoDelim)
	// Check for pre-existing grouping labels:
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "job" {
					return fmt.Errorf("pushed metric %s (%s) already contains a job label", mf.GetName(), m)
				}
				if _, ok := grouping[l.GetName()]; ok {
					return fmt.Errorf(
						"pushed metric %s (%s) already contains grouping label %s",
						mf.GetName(), m, l.GetName(),
					)
				}
			}
		}
		enc.Encode(mf)
	}
	req, err := http.NewRequest(method, pushURL, buf)
	if err != nil {
		return err
	}
	req.Header.Set(contentTypeHeader, string(expfmt.FmtProtoDelim))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		body, _ := ioutil.ReadAll(resp.Body) // Ignore any further error as this is for an error message only.
		return fmt.Errorf("unexpected status code %d while pushing to %s: %s", resp.StatusCode, pushURL, body)
	}
	return nil
}
