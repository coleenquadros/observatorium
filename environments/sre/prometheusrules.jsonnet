local thanos = (import 'thanos-mixin/mixin.libsonnet');
local thanosReceiveController = (import 'thanos-receive-controller-mixin/mixin.libsonnet');
local slo = import 'slo-libsonnet/slo.libsonnet';
local observatoriumSLOs = import '../../slos.libsonnet';

{
  'observatorium-slos-stage.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-slos-stage',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    spec+: {
      groups+: [
        {
          name: 'observatorium.slo.rules',
          rules:
            local rangeQueryErrors = observatoriumSLOs.observatoriumRangeQuery.errors {
              selectors+: ['namespace="telemeter-stage"'],
            };
            slo.errorburn(rangeQueryErrors).recordingrules +
            slo.errorburn(rangeQueryErrors).alerts,
        } + {
          // Override all severity to info for now
          rules: [
            r {
              labels+: {
                [if std.objectHas(r.labels, 'severity') then 'severity']: 'info',
              },
            }
            for r in super.rules
          ],
        },
      ],
    },
  },
  'observatorium-thanos-stage.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-stage',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos + thanosReceiveController {
      _config+:: {
        thanosQuerierJobPrefix: 'thanos-querier',
        thanosStoreJobPrefix: 'thanos-store',
        thanosReceiveJobPrefix: 'thanos-receive',
        thanosCompactJobPrefix: 'thanos-compactor',
        thanosReceiveControllerJobPrefix: 'thanos-receive-controller',

        thanosQuerierSelector: 'job=~"%s.*", namespace="telemeter-stage"' % self.thanosQuerierJobPrefix,
        thanosStoreSelector: 'job=~"%s.*", namespace="telemeter-stage"' % self.thanosStoreJobPrefix,
        thanosReceiveSelector: 'job=~"%s.*", namespace="telemeter-stage"' % self.thanosReceiveJobPrefix,
        thanosCompactSelector: 'job=~"%s.*", namespace="telemeter-stage"' % self.thanosCompactJobPrefix,
        thanosReceiveControllerSelector: 'job=~"%s.*",namespace="telemeter-stage"' % self.thanosReceiveControllerJobPrefix,

      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },

    spec: alerts.prometheusAlerts,
  },
  'observatorium-thanos-production.prometheusrules': {
    apiVersion: 'monitoring.coreos.com/v1',
    kind: 'PrometheusRule',
    metadata: {
      name: 'observatorium-thanos-production',
      labels: {
        prometheus: 'app-sre',
        role: 'alert-rules',
      },
    },
    local alerts = thanos + thanosReceiveController {
      _config+:: {
        thanosQuerierJobPrefix: 'thanos-querier',
        thanosStoreJobPrefix: 'thanos-store',
        thanosReceiveJobPrefix: 'thanos-receive',
        thanosCompactJobPrefix: 'thanos-compactor',
        thanosReceiveControllerJobPrefix: 'thanos-receive-controller',

        thanosQuerierSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosQuerierJobPrefix,
        thanosStoreSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosStoreJobPrefix,
        thanosReceiveSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosReceiveJobPrefix,
        thanosCompactSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosCompactJobPrefix,
        thanosReceiveControllerSelector: 'job=~"%s.*",namespace="telemeter-production"' % self.thanosReceiveControllerJobPrefix,

      },
    } + {
      prometheusAlerts+:: {
        groups:
          std.filter(
            function(ruleGroup) ruleGroup.name != 'thanos-sidecar.rules',
            super.groups,
          ),
      },
    },

    spec: alerts.prometheusAlerts,
  },
}
