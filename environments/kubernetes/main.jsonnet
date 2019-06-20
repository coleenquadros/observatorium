local app =
  (import 'kube-thanos.libsonnet');

{ ['thanos-querier-' + name]: app.thanos.querier[name] for name in std.objectFields(app.thanos.querier) } +
{ ['thanos-receive-' + name]: app.thanos.receive[name] for name in std.objectFields(app.thanos.receive) } +
{ ['thanos-store-' + name]: app.thanos.store[name] for name in std.objectFields(app.thanos.store) }
