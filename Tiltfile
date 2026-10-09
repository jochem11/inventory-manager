# Inventory manager — local development on Kubernetes (infra/development).
# Staging and prod are built and deployed by the pipelines.

### K8s Config ###

k8s_yaml('./infra/development/k8s/app-config.yaml')

# Credentials and local settings live in .env (never committed; see
# .env.example). They go into the cluster as the Secret app-secrets, which the
# deployments read. Editing .env reloads this; restart a service to pick it up.
def read_env(path):
  if not os.path.exists(path):
    fail('%s is missing: run `cp .env.example .env` and fill it in.' % path)
  values = {}
  for line in str(read_file(path)).splitlines():
    line = line.strip()
    if not line or line.startswith('#') or '=' not in line:
      continue
    key, value = line.split('=', 1)
    values[key.strip()] = value.strip()
  return values

k8s_yaml(encode_yaml({
  'apiVersion': 'v1',
  'kind': 'Secret',
  'metadata': {'name': 'app-secrets'},
  'type': 'Opaque',
  'stringData': read_env('.env'),
}))

### End of K8s Config ###


### Tracing ###

k8s_yaml('./infra/development/k8s/jaeger.yaml')
# Jaeger UI at http://localhost:16686.
k8s_resource('jaeger', port_forwards=16686, labels='observability')

### End of Tracing ###


### Kafka ###

k8s_yaml(['./infra/development/k8s/kafka.yaml', './infra/development/k8s/kafka-ui.yaml'])
# Services in the cluster use kafka:9092; `go run` on your machine uses 127.0.0.1:9094.
k8s_resource('kafka', port_forwards=['9094:9094'], labels='infra')
k8s_resource('kafka-topics', resource_deps=['kafka'], labels='infra')
# Kafka UI at http://localhost:8080. Off by default to save resources: start
# it with the ▶ button on kafka-ui in the Tilt UI.
k8s_resource('kafka-ui', port_forwards=['8080:8080'], resource_deps=['kafka'], labels='infra',
             auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL)

### End of Kafka ###


### Web Frontend ###

web_live_update = [
  # Config changes need a fresh Vite process.
  fall_back_on(['./web/vite.config.ts', './web/tsconfig.json']),
  sync('./web/package.json', '/app/package.json'),
  sync('./web/bun.lock', '/app/bun.lock'),
  sync('./web/src', '/app/src'),
  sync('./web/public', '/app/public'),
  run('bun install --frozen-lockfile', trigger=['./web/package.json', './web/bun.lock']),
]

if os.name == 'nt':
  custom_build(
    'inventory-manager/web',
    'infra\\development\\docker\\web-build.bat',
    deps=['./web', './infra/development/docker/web.Dockerfile'],
    live_update=web_live_update,
  )
else:
  docker_build(
    'inventory-manager/web',
    '.',
    dockerfile='./infra/development/docker/web.Dockerfile',
    only=['./web'],
    ignore=['web/docs', 'web/README.md'],
    live_update=web_live_update,
  )

k8s_yaml('./infra/development/k8s/web-deployment.yaml')
k8s_resource('web', port_forwards=3000, labels="frontend")

### End of Web Frontend ###


### Go services ###

# The Go services are compiled on your machine, not inside minikube: that's
# seconds instead of minutes, with your Go build cache. The binary lands in
# build/<service> and go-service.Dockerfile only copies it into an image.
# GOARCH is your machine's, which is also minikube's (both arm64 on an M-series Mac).
GOARCH = str(local('go env GOARCH', quiet=True)).strip()

def go_service(name, port, deps=[], ignore=[], label=None):
  """A Go service: compiled on your machine, put in an image, and deployed
  with infra/development/k8s/<name>-deployment.yaml, port-forwarded to port.
  It starts after its compile step and the resources in deps; ignore lists
  files in its folder that don't affect the binary."""
  local_resource(
    name + '-compile',
    'cd services/%s && GOWORK=off CGO_ENABLED=0 GOOS=linux GOARCH=%s go build -o ../../build/%s ./cmd' % (name, GOARCH, name),
    deps=['./shared', './services/' + name],
    ignore=['services/%s/%s' % (name, path) for path in ['Makefile', 'Readme.md', 'docs'] + ignore] + ['**/*_test.go'],
    labels=[label or name],
  )
  docker_build(
    'inventory-manager/' + name,
    './build',
    dockerfile='./infra/development/docker/go-service.Dockerfile',
    only=['./' + name],
    build_args={'BINARY': name},
  )
  k8s_yaml('./infra/development/k8s/%s-deployment.yaml' % name)
  k8s_resource(name, port_forwards=port, labels=[label or name], resource_deps=[name + '-compile'] + deps)

# The services use the MySQL on your machine and create their databases on
# first start.
go_service('user-service', 50051, deps=['kafka-topics'])
go_service('auth-service', 50052, deps=['kafka-topics'])
go_service('item-service', 50053)
# Playground at http://localhost:4000/graphql. Its schema is compiled into the
# generated code, so the schema files aren't needed in the binary.
go_service('graphql-gateway', 4000, deps=['user-service', 'auth-service', 'item-service'],
           ignore=['gqlgen.yml', 'schema'], label='gateway')

### End of Go services ###
