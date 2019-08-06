# Jolokia2 Input Plugins (Execute Method)

The [Jolokia](http://jolokia.org) _agent_ and _proxy_ input plugins collect JMX metrics from an HTTP endpoint using Jolokia's [JSON-over-HTTP protocol](https://jolokia.org/reference/html/protocol.html).

#### Note

The `Gather` method in `gatherer.go` is designed for the specific use case of ExecutionDetailQueue in the monitoring project. For generalized use cases, the `Gather` method need to be changed accordingly to collect the target metrics.

### Configuration:

#### Jolokia Agent Configuration

The `jolokia2_exec_agent` input plugin reads JMX metrics from one or more [Jolokia agent](https://jolokia.org/agent/jvm.html) REST endpoints.

Taking the specific use case of ExecutionDetailQueue in the monitoring project for example:

```toml
[[inputs.jolokia2_exec_agent]]
  urls = ["http://agent:8080/jolokia"]

  [[inputs.jolokia2_exec_agent.metric]]
    name      = "monitor_execution_details"
    mbean     = "com.intuit.platform.fdp.monitoring:type=ExecutionDetailQueue"
    operation = "poll"
    # arguments = []
    # repeatMetric = true
    # repeatTime = 10
```

Optionally, specify TLS options for communicating with agents:

```toml
[[inputs.jolokia2_exec_agent]]
  urls = ["https://agent:8080/jolokia"]
  tls_ca   = "/var/private/ca.pem"
  tls_cert = "/var/private/client.pem"
  tls_key  = "/var/private/client-key.pem"
  #insecure_skip_verify = false

  [[inputs.jolokia2_exec_agent.metric]]
    name      = "monitor_execution_details"
    mbean     = "com.intuit.platform.fdp.monitoring:type=ExecutionDetailQueue"
    operation = "poll"
    # arguments = []
    # repeatMetric = true
    # repeatTime = 10
```

#### Jolokia Proxy Configuration

The `jolokia2_exec_proxy` input plugin reads JMX metrics from one or more _targets_ by interacting with a [Jolokia proxy](https://jolokia.org/features/proxy.html) REST endpoint.

```toml
[[inputs.jolokia2_exec_proxy]]
  url = "http://proxy:8080/jolokia"

  #default_target_username = ""
  #default_target_password = ""
  [[inputs.jolokia2_exec_proxy.target]]
    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
    # username = ""
    # password = ""

  [[inputs.jolokia2_exec_proxy.metric]]
    name      = "monitor_execution_details"
    mbean     = "com.intuit.platform.fdp.monitoring:type=ExecutionDetailQueue"
    operation = "poll"
    # arguments = []
    # repeatMetric = true
    # repeatTime = 10
```

Optionally, specify TLS options for communicating with proxies:

```toml
[[inputs.jolokia2_exec_proxy]]
  url = "https://proxy:8080/jolokia"

  tls_ca   = "/var/private/ca.pem"
  tls_cert = "/var/private/client.pem"
  tls_key  = "/var/private/client-key.pem"
  #insecure_skip_verify = false

  #default_target_username = ""
  #default_target_password = ""
  [[inputs.jolokia2_exec_proxy.target]]
    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
    # username = ""
    # password = ""

  [[inputs.jolokia2_exec_proxy.metric]]
    name      = "monitor_execution_details"
    mbean     = "com.intuit.platform.fdp.monitoring:type=ExecutionDetailQueue"
    operation = "poll"
    # arguments = []
    # repeatMetric = true
    # repeatTime = 10
```

#### Jolokia Metric Configuration

Each `metric` declaration generates a Jolokia request to fetch telemetry from a JMX MBean.

| Key            | Required | Description |
|----------------|----------|-------------|
| `mbean`        | yes      | The object name of a JMX MBean. MBean property-key values can contain a wildcard `*`, allowing you to fetch multiple MBeans with one declaration. |
| `operation`    | yes      | The operation to execute on the given mBean. |
| `arguments`    | no       | An array of arguments for invoking this operation on the given mBean. |
| `repeatMetric` | yes      | A bool value to determine if we need to repeat the operation multiple times on this specific metric. |
| `repeatTime`   | --       | An int value to specify the times we want to repeat the operation on this metric. It is only required when `repeatMetric` is set to `true`. |
| `tag_keys`     | no       | A list of MBean property-key names to convert into tags. The property-key name becomes the tag name, while the property-key value becomes the tag value. |
| `tag_prefix`   | no       | A string to prepend to the tag names produced by this `metric` declaration. |
| `field_name`   | no       | A string to set as the name of the field produced by this metric; can contain substitutions. |
| `field_prefix` | no       | A string to prepend to the field names produced by this `metric` declaration; can contain substitutions. |

Both `jolokia2_exec_agent` and `jolokia2_exec_proxy` plugins support default configurations that apply to every `metric` declaration.

| Key                       | Default Value | Description |
|---------------------------|---------------|-------------|
| `default_field_separator` | `.`           | A character to use to join Mbean attributes when creating fields. |
| `default_field_prefix`    | _None_        | A string to prepend to the field names produced by all `metric` declarations. |
| `default_tag_prefix`      | _None_        | A string to prepend to the tag names produced by all `metric` declarations. |
