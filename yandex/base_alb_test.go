package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albDefaultSni = "tf-test-tls"
const albDefaultValidationContext = "tf-test-validation-context"
const albDefaultBackendWeight = "1"
const albDefaultPanicThreshold = "50"
const albDefaultLocalityPercent = "35"
const albDefaultTimeout = "300s"
const albDefaultInterval = "1m500s"
const albDefaultStrictLocality = "true"
const albDefaultServiceName = "true"
const albDefaultHttp2 = "true"
const albDefaultHost = "tf-test-host"
const albDefaultPath = "tf-test-path"
const albDefaultPort = "3"
const albDefaultSend = "tf-test-send"
const albDefaultReceive = "tf-test-receive"

func albBGDefaultALBValues() map[string]interface{} {
	return map[string]interface{}{
		"TGName":        acctest.RandomWithPrefix("tf-tg"),
		"BGName":        acctest.RandomWithPrefix("tf-bg"),
		"BGDescription": "alb-bg-descriprion",
		// "BaseTemplate":         testAccALBBaseTemplate(acctest.RandomWithPrefix("tf-instance")),
		"TlsSni":               albDefaultSni,
		"TlsValidationContext": albDefaultValidationContext,
		"BackendWeight":        albDefaultBackendWeight,
		"PanicThreshold":       albDefaultPanicThreshold,
		"LocalityPercent":      albDefaultLocalityPercent,
		"StrictLocality":       albDefaultStrictLocality,
		"Timeout":              albDefaultTimeout,
		"Interval":             albDefaultInterval,
		"ServiceName":          albDefaultServiceName,
		"Http2":                albDefaultHttp2,
		"Host":                 albDefaultHost,
		"Path":                 albDefaultPath,
		"Port":                 albDefaultPort,
		"Receive":              albDefaultReceive,
		"Send":                 albDefaultSend,
	}
}

func testAccCheckALBBackendGroupValues(bg *apploadbalancer.BackendGroup, expectedHttpBackends, expectedGrpcBackends bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (bg.GetHttp() != nil) != expectedHttpBackends {
			return fmt.Errorf("invalid presence or absence of http backend Application Backend Group %s", bg.Name)
		}

		if (bg.GetGrpc() != nil) != expectedGrpcBackends {
			return fmt.Errorf("invalid presence or absence of grpc backend Application Backend Group %s", bg.Name)
		}

		return nil
	}
}

func testAccALBGeneralTGTemplate(tgName, tgDesc, baseTemplate string, targetsCount int, isDataSource bool) string {
	targets := make([]string, targetsCount)
	for i := 1; i <= targetsCount; i++ {
		targets[i-1] = fmt.Sprintf("test-instance-%d", i)
	}
	return templateConfig(`
{{ if .IsDataSource }}
data "yandex_alb_target_group" "test-tg-ds" {
  name = yandex_alb_target_group.test-tg.name
}		
{{ end }}
resource "yandex_alb_target_group" "test-tg" {
  name        = "{{.TGName}}"
  description = "{{.TGDescription}}"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  {{range .Targets}}
  target {
    subnet_id  = yandex_vpc_subnet.test-subnet.id
    ip_address = yandex_compute_instance.{{.}}.network_interface.0.ip_address
  }		
  {{end}}
}

{{.BaseTemplate}}
		`,
		map[string]interface{}{
			"TGName":        tgName,
			"TGDescription": tgDesc,
			"BaseTemplate":  baseTemplate,
			"Targets":       targets,
			"IsDataSource":  isDataSource,
		},
	)
}

func testAccALBGeneralBGTemplate(ctx map[string]interface{}, isDataSource, isHttpBackend, isGrpcBackend, isHttpCheck, isGrpcCheck, isStreamCheck bool) string {
	ctx["IsHttpBackend"] = isHttpBackend
	ctx["IsGrpcBackend"] = isGrpcBackend
	ctx["IsBackend"] = isHttpBackend || isGrpcBackend
	ctx["IsHttpCheck"] = isHttpCheck
	ctx["IsGrpcCheck"] = isGrpcCheck
	ctx["IsStreamCheck"] = isStreamCheck
	ctx["BaseTemplate"] = testAccALBBaseTemplate(acctest.RandomWithPrefix("tf-instance"))
	ctx["IsDataSource"] = isDataSource
	return templateConfig(`
{{ if .IsDataSource }}
data "yandex_alb_backend_group" "test-bg-ds" {
  name = yandex_alb_backend_group.test-bg.name
}		
{{ end }}
resource "yandex_alb_backend_group" "test-bg" {
  name        = "{{.BGName}}"
  description = "{{.BGDescription}}"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  
  {{ if .IsHttpBackend }}
  http_backend {
    name = "test-http-backend"
    weight = {{.BackendWeight}}
    port = {{.Port}}
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      sni = "{{.TlsSni}}"
      validation_context {
        trusted_ca_bytes = "{{.TlsValidationContext}}"
      }
    }
    load_balancing_config {
      panic_threshold = {{.PanicThreshold}}
      locality_aware_routing_percent = {{.LocalityPercent}}
      strict_locality = {{.StrictLocality}}
    }

    {{ if .IsGrpcCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      grpc_healthcheck {
        service_name = "{{.ServiceName}}"
      }
    }
    {{end}}

    {{ if .IsStreamCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      stream_healthcheck {
        receive = "{{.Receive}}"
        send = "{{.Send}}"
      }
    }
    {{end}}

    {{ if .IsHttpCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      http_healthcheck {
        host = "{{.Host}}"
        path = "{{.Path}}"
        http2 = "{{.Http2}}"
      }
    }
    {{end}}

    http2 = "{{.Http2}}"
  }
  {{end}}

  {{ if .IsGrpcBackend }}
  grpc_backend {
    name = "test-grpc-backend"
    weight = {{.BackendWeight}}
    port = {{.Port}}
    target_group_ids = ["${yandex_alb_target_group.test-target-group.id}"]
    tls {
      sni = "{{.TlsSni}}"
      validation_context {
        trusted_ca_bytes = "{{.TlsValidationContext}}"
      }
    }
    load_balancing_config {
      panic_threshold = {{.PanicThreshold}}
      locality_aware_routing_percent = {{.LocalityPercent}}
      strict_locality = {{.StrictLocality}}
    }

    {{ if .IsGrpcCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      grpc_healthcheck {
        service_name = "{{.ServiceName}}"
      }
    }
    {{end}}

    {{ if .IsStreamCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      stream_healthcheck {
        receive = "{{.Receive}}"
        send = "{{.Send}}"
      }
    }
    {{end}}

    {{ if .IsHttpCheck }}
    healthcheck {
      timeout = "{{.Timeout}}"
      interval = "{{.Interval}}"
      http_healthcheck {
        host = "{{.Host}}"
        path = "{{.Path}}"
        http2 = "{{.Http2}}"
      }
    }
    {{end}}
  }
  {{end}}
}

{{ if or .IsHttpBackend .IsGrpcBackend }}
resource "yandex_alb_target_group" "test-target-group" {
  name		= "{{.TGName}}"

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address		= "${yandex_compute_instance.test-instance-1.network_interface.0.ip_address}"
  }

  target {
	subnet_id	= "${yandex_vpc_subnet.test-subnet.id}"
	ip_address		= "${yandex_compute_instance.test-instance-2.network_interface.0.ip_address}"
  }
}
{{ end }}

{{.BaseTemplate}}
		`,
		ctx,
	)
}

func testAccCheckALBTargetGroupValues(tg *apploadbalancer.TargetGroup, expectedInstanceNames []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		subnetIPMap, err := getSubnetIPMap(expectedInstanceNames)
		if err != nil {
			return err
		}

		if len(tg.GetTargets()) != len(expectedInstanceNames) {
			return fmt.Errorf("invalid count of targets in Application Target Group %s", tg.Name)
		}

		for _, t := range tg.GetTargets() {
			if addresses, ok := subnetIPMap[t.GetSubnetId()]; ok {
				addressExists := false
				for _, a := range addresses {
					if a == t.GetIpAddress() {
						addressExists = true
						break
					}
				}
				if !addressExists {
					return fmt.Errorf("invalid Target's Address %s in Application Target Group %s", t.GetIpAddress(), tg.Name)
				}
			} else {
				return fmt.Errorf("invalid Target's SubnetID %s in Application Target Group %s", t.GetSubnetId(), tg.Name)
			}
		}

		return nil
	}
}

func testAccALBBaseTemplate(instanceName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "test-image" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "test-instance-1" {
  name        = "%[1]s-1"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = data.yandex_compute_image.test-image.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.test-subnet.id
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_compute_instance" "test-instance-2" {
  name        = "%[1]s-2"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

  resources {
    cores         = 2
    core_fraction = 20
    memory        = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = data.yandex_compute_image.test-image.id
    }
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.test-subnet.id
  }

  scheduling_policy {
    preemptible = true
  }
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instanceName,
	)
}
